package user

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidPayload     = errors.New("invalid payload")
	ErrInvalidToken       = errors.New("invalid refresh token")
	ErrTokenRevoked       = errors.New("refresh token revoked")
	ErrTokenExpired       = errors.New("refresh token expired")
	ErrUnauthorized       = errors.New("unauthorized")
)
