package mapper

import (
	adminDto "github.com/Kooqoo22/JobJourney/backend/internal/admin/dto"
	authEntity "github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

func ToAdminUserResponse(u authEntity.User, tz string) adminDto.AdminUserResponse {
	resp := adminDto.AdminUserResponse{
		ID:          u.ID,
		Email:       u.Email,
		FullName:    u.FullName,
		AvatarURL:   u.AvatarURL,
		Timezone:    u.Timezone,
		IsVerified:  u.IsVerified,
		IsBanned:    u.IsBanned,
		BanReason:   u.BanReason,
		Role:        u.Role,
		CreatedAt:   utils.ToLocal(u.CreatedAt, tz).Format("2006-01-02T15:04:05Z07:00"),
	}
	if u.BannedAt != nil {
		s := utils.ToLocal(*u.BannedAt, tz).Format("2006-01-02T15:04:05Z07:00")
		resp.BannedAt = &s
	}
	if u.LastLoginAt != nil {
		s := utils.ToLocal(*u.LastLoginAt, tz).Format("2006-01-02T15:04:05Z07:00")
		resp.LastLoginAt = &s
	}
	return resp
}
