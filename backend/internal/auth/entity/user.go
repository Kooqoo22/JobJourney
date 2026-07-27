package entity

import "time"

type User struct {
	ID           int64      `db:"id"`
	Email        string     `db:"email"`
	PasswordHash *string    `db:"password_hash"`
	AuthProvider string     `db:"auth_provider"`
	FullName     string     `db:"full_name"`
	AvatarURL    *string    `db:"avatar_url"`
	Timezone     string     `db:"timezone"`
	IsVerified   bool       `db:"is_verified"`
	IsBanned     bool       `db:"is_banned"`
	BannedAt     *time.Time `db:"banned_at"`
	BanReason    *string    `db:"ban_reason"`
	Role         string     `db:"role"`
	LastLoginAt  *time.Time `db:"last_login_at"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}
