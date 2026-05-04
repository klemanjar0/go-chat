package auth

import (
	"context"

	domainauth "go-chat/internal/domain/auth"
	"go-chat/pkg/logger"
)

// Refresh validates and rotates a refresh token. If a previously rotated /
// revoked token is presented (token reuse), every refresh token for that user
// is revoked and ErrTokenRevoked is returned.
func (uc *UseCase) Refresh(ctx context.Context, refreshToken string, meta SessionMeta) (*AuthResult, error) {
	if refreshToken == "" {
		return nil, domainauth.ErrInvalidToken
	}

	now := uc.clock.Now()

	hash := uc.crypto.HashRefreshToken(refreshToken)
	stored, err := uc.refresh.GetByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	if stored.IsRevoked() {
		// Token reuse — assume the chain is compromised.
		logger.Warn("refresh token reuse detected", "user_id", stored.UserID, "token_id", stored.ID)
		if err := uc.refresh.RevokeAllForUser(ctx, stored.UserID); err != nil {
			logger.Error("revoke all on reuse failed", "err", err)
		}
		if err := uc.access.RevokeAllForUser(ctx, stored.UserID, uc.cfg.AccessTTL); err != nil {
			logger.Error("revoke access tokens on reuse failed", "err", err)
		}
		return nil, domainauth.ErrTokenRevoked
	}

	if stored.IsExpired(now) {
		return nil, domainauth.ErrTokenExpired
	}

	u, err := uc.users.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}

	newRaw, err := uc.crypto.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := now.Add(uc.cfg.RefreshTTL)
	if _, err := uc.refresh.Rotate(ctx, stored.ID, domainauth.CreateRefreshTokenInput{
		UserID:    stored.UserID,
		TokenHash: uc.crypto.HashRefreshToken(newRaw),
		ExpiresAt: expiresAt,
		UserAgent: meta.UserAgent,
		IP:        meta.IP,
	}); err != nil {
		return nil, err
	}

	access, err := uc.jwt.Issue(u.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User: u,
		Tokens: TokenPair{
			AccessToken:      access.Token,
			AccessExpiresAt:  access.ExpiresAt,
			RefreshToken:     newRaw,
			RefreshExpiresAt: expiresAt,
		},
	}, nil
}
