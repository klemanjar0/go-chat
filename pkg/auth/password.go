package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidPassword = errors.New("invalid password")

func HashPassword(password string, cost int) (string, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return ErrInvalidPassword
	}
	return nil
}
