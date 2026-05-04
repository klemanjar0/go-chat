package user

import "errors"

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUsernameTaken = errors.New("username already taken")
	ErrInvalidUserID = errors.New("invalid user id")
)
