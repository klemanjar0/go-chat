package auth

import "errors"

var (
	ErrInvalidPayload     = errors.New("invalid payload")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid refresh token")
	ErrTokenRevoked       = errors.New("refresh token revoked")
	ErrTokenExpired       = errors.New("refresh token expired")
	ErrUnauthorized       = errors.New("unauthorized")
)
