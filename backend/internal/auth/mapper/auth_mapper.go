package mapper

import (
	"time"

	"github.com/Kooqoo22/JobJourney/backend/internal/auth/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

func ToUserResponse(u entity.User) dto.UserResponse {
	return dto.UserResponse{
		ID:           u.ID,
		Email:        u.Email,
		FullName:     u.FullName,
		AvatarURL:    u.AvatarURL,
		Timezone:     u.Timezone,
		AuthProvider: u.AuthProvider,
		IsVerified:   u.IsVerified,
		Role:         u.Role,
		CreatedAt:    utils.ToLocal(u.CreatedAt, u.Timezone).Format(time.RFC3339),
	}
}

func ToTokenResponse(access, refresh string, expiresIn int64) dto.TokenResponse {
	return dto.TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}
}

func ToAuthResponse(u entity.User, access, refresh string, expiresIn int64) dto.AuthResponse {
	return dto.AuthResponse{
		User:   ToUserResponse(u),
		Tokens: ToTokenResponse(access, refresh, expiresIn),
	}
}
