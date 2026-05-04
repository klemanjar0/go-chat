package user

import (
	"context"

	domainauth "go-chat/internal/domain/auth"
	"go-chat/internal/domain/user"
)

const minPasswordLen = 8

func (uc *UseCase) UpdatePassword(ctx context.Context, userID, currentPassword, newPassword string) (*user.User, error) {
	if currentPassword == "" || len(newPassword) < minPasswordLen {
		return nil, domainauth.ErrInvalidPayload
	}

	u, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := uc.crypto.VerifyPassword(u.PasswordHash, currentPassword); err != nil {
		return nil, domainauth.ErrInvalidCredentials
	}

	hash, err := uc.crypto.HashPassword(newPassword, uc.bcryptCost)
	if err != nil {
		return nil, err
	}

	u.UpdatePassword(hash)
	return uc.users.Update(ctx, u)
}
