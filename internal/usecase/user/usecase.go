// Package user provides the user-domain use cases. Each public method on
// UseCase lives in its own file so that adding a new flow (e.g. update
// profile, delete account) is a self-contained change.
package user

import (
	"context"

	"go-chat/internal/domain/user"
)

// Store is the slice of user.Repository the user flows need.
type Store interface {
	GetByID(ctx context.Context, id string) (*user.User, error)
	Update(ctx context.Context, u *user.User) (*user.User, error)
}

// CryptoService is the slice of password primitives the user flows need.
// Production impl is pkg/auth.Service.
type CryptoService interface {
	HashPassword(password string, cost int) (string, error)
	VerifyPassword(hash, password string) error
}

type UseCase struct {
	users      Store
	crypto     CryptoService
	bcryptCost int
}

func NewUseCase(users Store, crypto CryptoService, bcryptCost int) *UseCase {
	return &UseCase{users: users, crypto: crypto, bcryptCost: bcryptCost}
}
