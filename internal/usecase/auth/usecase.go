// Package auth provides the authentication use cases. Each public method on
// UseCase lives in its own file so that adding a new flow (e.g. password
// reset, MFA) is a self-contained change.
package auth

import (
	"context"
	"net/netip"
	"time"

	"go-chat/internal/config"
	domainauth "go-chat/internal/domain/auth"
	"go-chat/internal/domain/user"
	pkgauth "go-chat/pkg/auth"
)

// UserStore is the slice of user.Repository the auth flows need.
type UserStore interface {
	Create(ctx context.Context, in user.CreateInput) (*user.User, error)
	GetByID(ctx context.Context, id string) (*user.User, error)
	GetByUsername(ctx context.Context, username string) (*user.User, error)
}

// RefreshTokenStore is the slice of auth.RefreshTokenRepository the auth flows need.
type RefreshTokenStore interface {
	Create(ctx context.Context, in domainauth.CreateRefreshTokenInput) (*domainauth.RefreshToken, error)
	GetByHash(ctx context.Context, hash string) (*domainauth.RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string) error
	Rotate(ctx context.Context, oldID string, newToken domainauth.CreateRefreshTokenInput) (*domainauth.RefreshToken, error)
}

// AccessTokenStore is the slice of auth.AccessTokenStore the auth flows need.
type AccessTokenStore interface {
	RevokeJTI(ctx context.Context, jti string, expiresAt time.Time) error
	IsJTIRevoked(ctx context.Context, jti string) (bool, error)
	RevokeAllForUser(ctx context.Context, userID string, accessTTL time.Duration) error
	RevokedBefore(ctx context.Context, userID string) (int64, error)
}

// JWTService issues and verifies short-lived access tokens.
type JWTService interface {
	Issue(userID string) (*pkgauth.IssuedAccessToken, error)
	Verify(token string) (*pkgauth.AccessClaims, error)
}

type SessionMeta struct {
	UserAgent *string
	IP        *netip.Addr
}

type TokenPair struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}

type AuthResult struct {
	User   *user.User
	Tokens TokenPair
}

type UseCase struct {
	cfg     config.Auth
	users   UserStore
	refresh RefreshTokenStore
	access  AccessTokenStore
	jwt     JWTService
}

func NewUseCase(
	cfg config.Auth,
	users UserStore,
	refresh RefreshTokenStore,
	access AccessTokenStore,
	jwt JWTService,
) *UseCase {
	return &UseCase{
		cfg:     cfg,
		users:   users,
		refresh: refresh,
		access:  access,
		jwt:     jwt,
	}
}

// issueTokenPair signs a fresh access token and persists a new refresh token.
// Used by Register, Login.
func (uc *UseCase) issueTokenPair(ctx context.Context, userID string, meta SessionMeta) (*TokenPair, error) {
	access, err := uc.jwt.Issue(userID)
	if err != nil {
		return nil, err
	}

	rawRefresh, err := pkgauth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(uc.cfg.RefreshTTL)
	if _, err := uc.refresh.Create(ctx, domainauth.CreateRefreshTokenInput{
		UserID:    userID,
		TokenHash: pkgauth.HashRefreshToken(rawRefresh),
		ExpiresAt: expiresAt,
		UserAgent: meta.UserAgent,
		IP:        meta.IP,
	}); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:      access.Token,
		AccessExpiresAt:  access.ExpiresAt,
		RefreshToken:     rawRefresh,
		RefreshExpiresAt: expiresAt,
	}, nil
}
