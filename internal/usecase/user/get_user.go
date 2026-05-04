package user

import (
	"context"

	"go-chat/internal/domain/user"
)

func (uc *UseCase) GetByID(ctx context.Context, id string) (*user.User, error) {
	return uc.users.GetByID(ctx, id)
}
