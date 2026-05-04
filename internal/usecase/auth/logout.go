package auth

import (
	"context"
	"errors"
	"time"

	domainauth "go-chat/internal/domain/auth"
	pkgauth "go-chat/pkg/auth"
)

// Logout revokes the supplied refresh token and adds the access-token jti to
// the denylist until its natural expiry.
func (uc *UseCase) Logout(ctx context.Context, refreshToken, accessJTI string, accessExpiresAt time.Time) error {
	if refreshToken != "" {
		hash := pkgauth.HashRefreshToken(refreshToken)
		stored, err := uc.refresh.GetByHash(ctx, hash)
		if err == nil {
			if err := uc.refresh.Revoke(ctx, stored.ID); err != nil {
				return err
			}
		} else if !errors.Is(err, domainauth.ErrInvalidToken) {
			return err
		}
	}

	if accessJTI != "" {
		if err := uc.access.RevokeJTI(ctx, accessJTI, accessExpiresAt); err != nil {
			return err
		}
	}
	return nil
}
