package usecase

import (
	"context"

	adminDto "github.com/Kooqoo22/JobJourney/backend/internal/admin/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/admin/mapper"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

type AdminUsecase struct {
	repo AdminRepoIface
}

func New(repo AdminRepoIface) *AdminUsecase {
	return &AdminUsecase{repo: repo}
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
