package user

import (
	"context"

	"go-chat/internal/domain/user"
)

func (uc *UseCase) UpdateProfile(ctx context.Context, userID, firstName, lastName string) (*user.User, error) {
	u, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	u.UpdateProfile(firstName, lastName)
	return uc.users.Update(ctx, u)
}
