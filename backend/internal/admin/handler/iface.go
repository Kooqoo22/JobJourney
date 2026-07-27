package handler

import (
	"context"

	adminDto "github.com/Kooqoo22/JobJourney/backend/internal/admin/dto"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

type AdminUsecaseIface interface {
	ListUsers(ctx context.Context, q adminDto.ListUsersQuery, userTZ string) ([]adminDto.AdminUserResponse, utils.PageMeta, error)
	BanUser(ctx context.Context, adminID, userID int64, reason *string) error
	UnbanUser(ctx context.Context, adminID, userID int64) error
	DeleteUser(ctx context.Context, adminID, userID int64) error
}
