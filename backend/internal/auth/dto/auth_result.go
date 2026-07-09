package dto

import "github.com/Kooqoo22/JobJourney/backend/internal/auth/entity"

type AuthResult struct {
	User         entity.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}
