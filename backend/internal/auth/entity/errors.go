package entity

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrEmailTokenNotFound   = errors.New("email token not found")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
)
