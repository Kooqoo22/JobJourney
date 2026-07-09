package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"unicode"

	"github.com/Kooqoo22/JobJourney/backend/internal/auth/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
	"github.com/Kooqoo22/JobJourney/backend/internal/auth/repository"
	"github.com/Kooqoo22/JobJourney/backend/internal/database"
	"github.com/Kooqoo22/JobJourney/backend/pkg/mailer"
	"github.com/Kooqoo22/JobJourney/backend/pkg/security"
	"github.com/Kooqoo22/JobJourney/backend/pkg/token"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

const genericAuthMessage = "email or password is incorrect"

type AuthUsecase struct {
	userRepo    *repository.UserRepository
	emailTokens *repository.EmailTokenRepository
	refresh     *repository.RefreshTokenRepository
	tx          *database.TxManager
	tokens      *token.Manager
	mail        mailer.Mailer

	frontendURL     string
	defaultTimezone string
	verifyTokenTTL  time.Duration
	resetTokenTTL   time.Duration
}

type Deps struct {
	UserRepo        *repository.UserRepository
	EmailTokens     *repository.EmailTokenRepository
	Refresh         *repository.RefreshTokenRepository
	Tx              *database.TxManager
	Tokens          *token.Manager
	Mailer          mailer.Mailer
	FrontendURL     string
	DefaultTimezone string
	VerifyTokenTTL  time.Duration
	ResetTokenTTL   time.Duration
}

func New(d Deps) *AuthUsecase {
	return &AuthUsecase{
		userRepo:        d.UserRepo,
		emailTokens:     d.EmailTokens,
		refresh:         d.Refresh,
		tx:              d.Tx,
		tokens:          d.Tokens,
		mail:            d.Mailer,
		frontendURL:     d.FrontendURL,
		defaultTimezone: d.DefaultTimezone,
		verifyTokenTTL:  d.VerifyTokenTTL,
		resetTokenTTL:   d.ResetTokenTTL,
	}
}

func (u *AuthUsecase) Register(ctx context.Context, req dto.RegisterRequest) (entity.User, error) {
	if fields := passwordStrengthErrors("password", req.Password); fields != nil {
		return entity.User{}, utils.ErrUnprocessable("password does not meet the requirements", fields)
	}

	tz := u.resolveTimezone(req.Timezone)
	if !isValidTimezone(tz) {
		return entity.User{}, utils.ErrUnprocessable("validation failed", []utils.FieldError{{Field: "timezone", Message: "must be a valid IANA timezone"}})
	}

	exists, err := u.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return entity.User{}, utils.ErrInternal(err)
	}
	if exists {
		return entity.User{}, utils.ErrConflict("email is already registered")
	}

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return entity.User{}, utils.ErrInternal(err)
	}

	user := entity.User{
		Email:        req.Email,
		PasswordHash: &hash,
		AuthProvider: "local",
		FullName:     req.FullName,
		Timezone:     tz,
		IsVerified:   false,
		Role:         "user",
	}

	rawToken, err := security.GenerateOpaqueToken()
	if err != nil {
		return entity.User{}, utils.ErrInternal(err)
	}

	err = u.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.userRepo.Insert(txCtx, &user); err != nil {
			if database.IsUniqueViolation(err) {
				return utils.ErrConflict("email is already registered")
			}
			return err
		}
		et := entity.EmailToken{
			UserID:    user.ID,
			Type:      "verify",
			TokenHash: security.HashToken(rawToken),
			ExpiresAt: time.Now().UTC().Add(u.verifyTokenTTL),
		}
		return u.emailTokens.Insert(txCtx, &et)
	})
	if err != nil {
		return entity.User{}, wrapInternal(err)
	}

	u.sendVerificationAsync(user.Email, user.FullName, rawToken)
	return user, nil
}

func (u *AuthUsecase) VerifyEmail(ctx context.Context, rawToken string) (entity.User, error) {
	hash := security.HashToken(rawToken)
	var verified entity.User

	err := u.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		et, err := u.emailTokens.GetActiveByHash(txCtx, hash, "verify")
		if err != nil {
			if errors.Is(err, entity.ErrEmailTokenNotFound) {
				return utils.ErrBadRequest("verification link is invalid")
			}
			return err
		}
		if et.UsedAt != nil {
			return utils.ErrBadRequest("verification link has already been used")
		}
		if et.ExpiresAt.Before(time.Now().UTC()) {
			return utils.ErrBadRequest("verification link has expired")
		}

		user, err := u.userRepo.GetByID(txCtx, et.UserID)
		if err != nil {
			if errors.Is(err, entity.ErrUserNotFound) {
				return utils.ErrNotFound("account not found")
			}
			return err
		}
		if user.IsBanned {
			return utils.ErrForbidden("your account has been disabled")
		}

		if !user.IsVerified {
			if err := u.emailTokens.MarkUsed(txCtx, et.ID); err != nil {
				return err
			}
			if err := u.userRepo.SetVerified(txCtx, user.ID); err != nil {
				return err
			}
			user.IsVerified = true
		}
		verified = user
		return nil
	})
	if err != nil {
		return entity.User{}, wrapInternal(err)
	}
	return verified, nil
}

func (u *AuthUsecase) ResendVerification(ctx context.Context, email string) error {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return nil
		}
		return utils.ErrInternal(err)
	}
	if user.AuthProvider != "local" || user.IsVerified || user.IsBanned {
		return nil
	}

	rawToken, err := security.GenerateOpaqueToken()
	if err != nil {
		return utils.ErrInternal(err)
	}

	err = u.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.emailTokens.InvalidateActive(txCtx, user.ID, "verify"); err != nil {
			return err
		}
		et := entity.EmailToken{
			UserID:    user.ID,
			Type:      "verify",
			TokenHash: security.HashToken(rawToken),
			ExpiresAt: time.Now().UTC().Add(u.verifyTokenTTL),
		}
		return u.emailTokens.Insert(txCtx, &et)
	})
	if err != nil {
		return wrapInternal(err)
	}

	u.sendVerificationAsync(user.Email, user.FullName, rawToken)
	return nil
}

func (u *AuthUsecase) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResult, error) {
	user, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return nil, utils.ErrUnauthorized(genericAuthMessage)
		}
		return nil, utils.ErrInternal(err)
	}
	if user.PasswordHash == nil || !security.CheckPassword(*user.PasswordHash, req.Password) {
		return nil, utils.ErrUnauthorized(genericAuthMessage)
	}
	if user.IsBanned {
		return nil, utils.ErrForbidden("your account has been disabled")
	}

	result, err := u.issueSession(ctx, user, true)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (u *AuthUsecase) Refresh(ctx context.Context, rawToken string) (*dto.AuthResult, error) {
	hash := security.HashToken(rawToken)

	stored, err := u.refresh.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, entity.ErrRefreshTokenNotFound) {
			return nil, utils.ErrUnauthorized("invalid refresh token")
		}
		return nil, utils.ErrInternal(err)
	}

	if stored.RevokedAt != nil {
		if rerr := u.refresh.RevokeAllByUser(ctx, stored.UserID); rerr != nil {
			return nil, utils.ErrInternal(rerr)
		}
		return nil, utils.ErrUnauthorized("refresh token has been revoked")
	}
	if stored.ExpiresAt.Before(time.Now().UTC()) {
		if rerr := u.refresh.Revoke(ctx, stored.ID); rerr != nil {
			return nil, utils.ErrInternal(rerr)
		}
		return nil, utils.ErrUnauthorized("refresh token has expired")
	}

	user, err := u.userRepo.GetByID(ctx, stored.UserID)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return nil, utils.ErrUnauthorized("invalid refresh token")
		}
		return nil, utils.ErrInternal(err)
	}
	if user.IsBanned {
		if rerr := u.refresh.RevokeAllByUser(ctx, user.ID); rerr != nil {
			return nil, utils.ErrInternal(rerr)
		}
		return nil, utils.ErrForbidden("your account has been disabled")
	}

	access, err := u.tokens.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	newRaw, err := security.GenerateOpaqueToken()
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	err = u.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.refresh.Revoke(txCtx, stored.ID); err != nil {
			return err
		}
		rt := entity.RefreshToken{
			UserID:    user.ID,
			TokenHash: security.HashToken(newRaw),
			ExpiresAt: time.Now().UTC().Add(u.tokens.RefreshTokenTTL()),
		}
		return u.refresh.Insert(txCtx, &rt)
	})
	if err != nil {
		return nil, wrapInternal(err)
	}

	return &dto.AuthResult{
		User:         user,
		AccessToken:  access,
		RefreshToken: newRaw,
		ExpiresIn:    int64(u.tokens.AccessTokenTTL().Seconds()),
	}, nil
}

func (u *AuthUsecase) Logout(ctx context.Context, rawToken string) error {
	if err := u.refresh.RevokeByHash(ctx, security.HashToken(rawToken)); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func (u *AuthUsecase) LogoutAll(ctx context.Context, userID int64) error {
	if err := u.refresh.RevokeAllByUser(ctx, userID); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func (u *AuthUsecase) ForgotPassword(ctx context.Context, email string) error {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return nil
		}
		return utils.ErrInternal(err)
	}
	if user.AuthProvider != "local" || user.PasswordHash == nil || user.IsBanned {
		return nil
	}

	rawToken, err := security.GenerateOpaqueToken()
	if err != nil {
		return utils.ErrInternal(err)
	}

	err = u.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.emailTokens.InvalidateActive(txCtx, user.ID, "reset"); err != nil {
			return err
		}
		et := entity.EmailToken{
			UserID:    user.ID,
			Type:      "reset",
			TokenHash: security.HashToken(rawToken),
			ExpiresAt: time.Now().UTC().Add(u.resetTokenTTL),
		}
		return u.emailTokens.Insert(txCtx, &et)
	})
	if err != nil {
		return wrapInternal(err)
	}

	u.sendPasswordResetAsync(user.Email, user.FullName, rawToken)
	return nil
}

func (u *AuthUsecase) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if fields := passwordStrengthErrors("new_password", newPassword); fields != nil {
		return utils.ErrUnprocessable("password does not meet the requirements", fields)
	}

	newHash, err := security.HashPassword(newPassword)
	if err != nil {
		return utils.ErrInternal(err)
	}
	hash := security.HashToken(rawToken)

	err = u.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		et, err := u.emailTokens.GetActiveByHash(txCtx, hash, "reset")
		if err != nil {
			if errors.Is(err, entity.ErrEmailTokenNotFound) {
				return utils.ErrBadRequest("reset link is invalid")
			}
			return err
		}
		if et.UsedAt != nil {
			return utils.ErrBadRequest("reset link has already been used")
		}
		if et.ExpiresAt.Before(time.Now().UTC()) {
			return utils.ErrBadRequest("reset link has expired")
		}

		user, err := u.userRepo.GetByID(txCtx, et.UserID)
		if err != nil {
			if errors.Is(err, entity.ErrUserNotFound) {
				return utils.ErrNotFound("account not found")
			}
			return err
		}
		if user.IsBanned {
			return utils.ErrForbidden("your account has been disabled")
		}

		if err := u.userRepo.UpdatePassword(txCtx, user.ID, newHash); err != nil {
			return err
		}
		if err := u.emailTokens.MarkUsed(txCtx, et.ID); err != nil {
			return err
		}
		return u.refresh.RevokeAllByUser(txCtx, user.ID)
	})
	if err != nil {
		return wrapInternal(err)
	}
	return nil
}

func (u *AuthUsecase) issueSession(ctx context.Context, user entity.User, touchLogin bool) (*dto.AuthResult, error) {
	access, err := u.tokens.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	rawRefresh, err := security.GenerateOpaqueToken()
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	err = u.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		rt := entity.RefreshToken{
			UserID:    user.ID,
			TokenHash: security.HashToken(rawRefresh),
			ExpiresAt: time.Now().UTC().Add(u.tokens.RefreshTokenTTL()),
		}
		if err := u.refresh.Insert(txCtx, &rt); err != nil {
			return err
		}
		if touchLogin {
			return u.userRepo.UpdateLastLogin(txCtx, user.ID)
		}
		return nil
	})
	if err != nil {
		return nil, wrapInternal(err)
	}

	return &dto.AuthResult{
		User:         user,
		AccessToken:  access,
		RefreshToken: rawRefresh,
		ExpiresIn:    int64(u.tokens.AccessTokenTTL().Seconds()),
	}, nil
}

func (u *AuthUsecase) resolveTimezone(tz string) string {
	if tz == "" {
		return u.defaultTimezone
	}
	return tz
}

func (u *AuthUsecase) sendVerificationAsync(email, name, rawToken string) {
	link := u.frontendURL + "/verify-email?token=" + rawToken
	go func() {
		if err := u.mail.SendVerificationEmail(email, name, link); err != nil {
			slog.Error("failed to send verification email", "error", err, "email", email)
		}
	}()
}

func (u *AuthUsecase) sendPasswordResetAsync(email, name, rawToken string) {
	link := u.frontendURL + "/reset-password?token=" + rawToken
	go func() {
		if err := u.mail.SendPasswordResetEmail(email, name, link); err != nil {
			slog.Error("failed to send password reset email", "error", err, "email", email)
		}
	}()
}

func wrapInternal(err error) error {
	var appErr *utils.AppError
	if errors.As(err, &appErr) {
		return err
	}
	return utils.ErrInternal(err)
}

func passwordStrengthErrors(field, pw string) []utils.FieldError {
	var hasLetter, hasDigit bool
	for _, r := range pw {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	var out []utils.FieldError
	if len([]rune(pw)) < 8 {
		out = append(out, utils.FieldError{Field: field, Message: "must be at least 8 characters"})
	}
	if !hasLetter || !hasDigit {
		out = append(out, utils.FieldError{Field: field, Message: "must contain both letters and numbers"})
	}
	return out
}

func isValidTimezone(tz string) bool {
	_, err := time.LoadLocation(tz)
	return err == nil
}
