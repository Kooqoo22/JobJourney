package dto

type ProfileResponse struct {
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
