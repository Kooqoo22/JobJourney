package dto

type UserResponse struct {
	ID           int64   `json:"id"`
	Email        string  `json:"email"`
	FullName     string  `json:"full_name"`
	AvatarURL    *string `json:"avatar_url"`
	Timezone     string  `json:"timezone"`
	AuthProvider string  `json:"auth_provider"`
	IsVerified   bool    `json:"is_verified"`
	Role         string  `json:"role"`
	CreatedAt    string  `json:"created_at"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type AuthResponse struct {
	User   UserResponse  `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}
