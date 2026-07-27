package dto

type AdminUserResponse struct {
	ID          int64   `json:"id"`
	Email       string  `json:"email"`
	FullName    string  `json:"full_name"`
	AvatarURL   *string `json:"avatar_url"`
	Timezone    string  `json:"timezone"`
	IsVerified  bool    `json:"is_verified"`
	IsBanned    bool    `json:"is_banned"`
	BanReason   *string `json:"ban_reason"`
	BannedAt    *string `json:"banned_at"`
	Role        string  `json:"role"`
	LastLoginAt *string `json:"last_login_at"`
	CreatedAt   string  `json:"created_at"`
}
