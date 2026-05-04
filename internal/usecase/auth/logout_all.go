package auth

import "context"

// LogoutAll revokes every refresh token for the user and stamps a per-user
// access-token watermark in Redis so all currently-issued access tokens are
// rejected by the middleware.
func (uc *UseCase) LogoutAll(ctx context.Context, userID string) error {
	if err := uc.refresh.RevokeAllForUser(ctx, userID); err != nil {
		return err
	}
	return uc.access.RevokeAllForUser(ctx, userID, uc.cfg.AccessTTL)
}
