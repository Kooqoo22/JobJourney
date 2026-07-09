package handler

import (
	"context"

	"github.com/Kooqoo22/JobJourney/backend/internal/auth/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
)

type AuthUsecaseIface interface {
	Register(ctx context.Context, req dto.RegisterRequest) (entity.User, error)
	VerifyEmail(ctx context.Context, rawToken string) (entity.User, error)
	ResendVerification(ctx context.Context, email string) error
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResult, error)
	Refresh(ctx context.Context, rawToken string) (*dto.AuthResult, error)
	Logout(ctx context.Context, rawToken string) error
	LogoutAll(ctx context.Context, userID int64) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, rawToken, newPassword string) error
}
