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

// CryptoService is the slice of password / refresh-token primitives the auth
// use case depends on. Production impl is pkg/auth.Service.
type CryptoService interface {
	HashPassword(password string, cost int) (string, error)
	VerifyPassword(hash, password string) error
	GenerateRefreshToken() (string, error)
	HashRefreshToken(token string) string
}

// Clock returns the current time. Injected so flows that compute expiries or
// compare timestamps can be tested with a fixed instant.
type Clock interface {
	Now() time.Time
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
	crypto  CryptoService
	clock   Clock
}

func NewUseCase(
	cfg config.Auth,
	users UserStore,
	refresh RefreshTokenStore,
	access AccessTokenStore,
	jwt JWTService,
	crypto CryptoService,
	clock Clock,
) *UseCase {
	return &UseCase{
		cfg:     cfg,
		users:   users,
		refresh: refresh,
		access:  access,
		jwt:     jwt,
		crypto:  crypto,
		clock:   clock,
	}
}

// issueTokenPair signs a fresh access token and persists a new refresh token.
// Used by Register and Login.
func (uc *UseCase) issueTokenPair(ctx context.Context, userID string, meta SessionMeta) (*TokenPair, error) {
	access, err := uc.jwt.Issue(userID)
	if err != nil {
		return nil, err
	}

	rawRefresh, err := uc.crypto.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := uc.clock.Now().Add(uc.cfg.RefreshTTL)
	if _, err := uc.refresh.Create(ctx, domainauth.CreateRefreshTokenInput{
		UserID:    userID,
		TokenHash: uc.crypto.HashRefreshToken(rawRefresh),
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
