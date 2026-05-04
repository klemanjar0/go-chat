package user

import (
	"context"

	"go-chat/internal/domain/user"
)

func (uc *UseCase) SetAvatar(ctx context.Context, userID string, avatarURL *string) (*user.User, error) {
	u, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	u.SetAvatar(avatarURL)
	return uc.users.Update(ctx, u)
}
