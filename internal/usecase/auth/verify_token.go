package auth

import (
	"context"

	domainauth "go-chat/internal/domain/auth"
	pkgauth "go-chat/pkg/auth"
)

// VerifyAccessToken verifies the JWT signature/expiry, then checks the Redis
// denylist (jti) and the user's revoked-before watermark.
func (uc *UseCase) VerifyAccessToken(ctx context.Context, tokenStr string) (*pkgauth.AccessClaims, error) {
	claims, err := uc.jwt.Verify(tokenStr)
	if err != nil {
		return nil, err
	}

	revoked, err := uc.access.IsJTIRevoked(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, domainauth.ErrUnauthorized
	}

	watermark, err := uc.access.RevokedBefore(ctx, claims.Subject)
	if err != nil {
		return nil, err
	}
	if watermark > 0 && claims.IssuedAt != nil && claims.IssuedAt.Unix() < watermark {
		return nil, domainauth.ErrUnauthorized
	}
	return claims, nil
}
