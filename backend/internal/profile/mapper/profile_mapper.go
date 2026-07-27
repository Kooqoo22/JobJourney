package mapper

import (
	"time"

	"github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
	"github.com/Kooqoo22/JobJourney/backend/internal/profile/dto"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

func ToProfileResponse(u entity.User) dto.ProfileResponse {
	return dto.ProfileResponse{
		ID:         u.ID,
		Email:      u.Email,
		FullName:   u.FullName,
		AvatarURL:  u.AvatarURL,
		Timezone:   u.Timezone,
		IsVerified: u.IsVerified,
		Role:       u.Role,
		CreatedAt:  utils.ToLocal(u.CreatedAt, u.Timezone).Format(time.RFC3339),
	}
}
