package usecase

import (
	"context"

	adminDto "github.com/Kooqoo22/JobJourney/backend/internal/admin/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/admin/mapper"
	authEntity "github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

type AdminUsecase struct {
	repo AdminRepoIface
	tx   TxManagerIface
}

func New(repo AdminRepoIface, tx TxManagerIface) *AdminUsecase {
	return &AdminUsecase{repo: repo, tx: tx}
}

func (u *AdminUsecase) ListUsers(ctx context.Context, q adminDto.ListUsersQuery, userTZ string) ([]adminDto.AdminUserResponse, utils.PageMeta, error) {
	page := utils.NormalizePage(q.Page)
	limit := utils.NormalizeLimit(q.Limit)
	offset := (page - 1) * limit

	users, total, err := u.repo.ListUsers(ctx, q.Q, q.Status, offset, limit)
	if err != nil {
		return nil, utils.PageMeta{}, utils.ErrInternal(err)
	}

	responses := make([]adminDto.AdminUserResponse, len(users))
	for i, user := range users {
		responses[i] = mapper.ToAdminUserResponse(user, userTZ)
	}
	return responses, utils.NewPageMeta(total, page, limit), nil
}

func (u *AdminUsecase) BanUser(ctx context.Context, adminID, userID int64, reason *string) error {
	if adminID == userID {
		return utils.ErrUnprocessable("validation failed", []utils.FieldError{
			{Field: "id", Message: "cannot ban yourself"},
		})
	}
	target, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		if isNotFound(err) {
			return utils.ErrNotFound("user not found")
		}
		return utils.ErrInternal(err)
	}
	if target.IsBanned {
		return utils.ErrConflict("user is already banned")
	}
	return u.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.repo.BanUser(txCtx, userID, reason); err != nil {
			return utils.ErrInternal(err)
		}
		if err := u.repo.RevokeAllTokensByUserID(txCtx, userID); err != nil {
			return utils.ErrInternal(err)
		}
		return nil
	})
}

func (u *AdminUsecase) UnbanUser(ctx context.Context, adminID, userID int64) error {
	if adminID == userID {
		return utils.ErrUnprocessable("validation failed", []utils.FieldError{
			{Field: "id", Message: "cannot unban yourself"},
		})
	}
	target, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		if isNotFound(err) {
			return utils.ErrNotFound("user not found")
		}
		return utils.ErrInternal(err)
	}
	if !target.IsBanned {
		return utils.ErrConflict("user is not banned")
	}
	if err := u.repo.UnbanUser(ctx, userID); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func (u *AdminUsecase) DeleteUser(ctx context.Context, adminID, userID int64) error {
	if adminID == userID {
		return utils.ErrUnprocessable("validation failed", []utils.FieldError{
			{Field: "id", Message: "cannot delete yourself"},
		})
	}
	if _, err := u.repo.GetUserByID(ctx, userID); err != nil {
		if isNotFound(err) {
			return utils.ErrNotFound("user not found")
		}
		return utils.ErrInternal(err)
	}
	return u.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.repo.SoftDeleteUserData(txCtx, userID); err != nil {
			return utils.ErrInternal(err)
		}
		if err := u.repo.SoftDeleteUser(txCtx, userID); err != nil {
			return utils.ErrInternal(err)
		}
		return nil
	})
}

func isNotFound(err error) bool {
	return err == authEntity.ErrUserNotFound
}
