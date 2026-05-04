package user

import (
	"context"
	"errors"
	"net/netip"
	"strings"
	"time"

	"go-chat/internal/configuration"
	"go-chat/pkg/auth"
	"go-chat/pkg/logger"

	"github.com/jackc/pgx/v5/pgconn"
)

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
	User   *User
	Tokens TokenPair
}

type UseCase struct {
	cfg   *configuration.AuthConfig
	repo  *Repository
	store *TokenStore
	jwt   *auth.JWTIssuer
}

func NewUseCase(cfg *configuration.AuthConfig, repo *Repository, store *TokenStore) *UseCase {
	return &UseCase{
		cfg:   cfg,
		repo:  repo,
		store: store,
		jwt:   auth.NewJWTIssuer(cfg.JWTSecret, cfg.JWTIssuer, cfg.AccessTokenTTL),
	}
}

type RegisterInput struct {
	Username  string
	Password  string
	FirstName string
	LastName  string
	Phone     string
	AvatarURL *string
}

func (uc *UseCase) Register(ctx context.Context, in RegisterInput, meta SessionMeta) (*AuthResult, error) {
	in.Username = strings.TrimSpace(in.Username)
	if in.Username == "" || len(in.Password) < 8 {
		return nil, ErrInvalidPayload
	}

	hash, err := auth.HashPassword(in.Password, uc.cfg.BcryptCost)
	if err != nil {
		return nil, err
	}

	created, err := uc.repo.CreateUser(ctx, CreateUserInput{
		Username:     in.Username,
		PasswordHash: hash,
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		Phone:        in.Phone,
		AvatarURL:    in.AvatarURL,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUsernameTaken
		}
		return nil, err
	}

	tokens, err := uc.issueTokenPair(ctx, created.ID, meta)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: created, Tokens: *tokens}, nil
}

func (uc *UseCase) Login(ctx context.Context, username, password string, meta SessionMeta) (*AuthResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	u, err := uc.repo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := auth.VerifyPassword(u.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	tokens, err := uc.issueTokenPair(ctx, u.ID, meta)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: u, Tokens: *tokens}, nil
}

// Refresh validates and rotates a refresh token. If a previously rotated/revoked
// token is presented (token reuse), every refresh token for that user is
// revoked and ErrTokenRevoked is returned.
func (uc *UseCase) Refresh(ctx context.Context, refreshToken string, meta SessionMeta) (*AuthResult, error) {
	if refreshToken == "" {
		return nil, ErrInvalidToken
	}

	hash := auth.HashRefreshToken(refreshToken)
	stored, err := uc.repo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	if stored.IsRevoked() {
		// Token reuse — assume the chain is compromised.
		logger.Warn("refresh token reuse detected", "user_id", stored.UserID, "token_id", stored.ID)
		if err := uc.repo.RevokeAllUserRefreshTokens(ctx, stored.UserID); err != nil {
			logger.Error("revoke all on reuse failed", "err", err)
		}
		if err := uc.store.RevokeAllUserAccessTokens(ctx, stored.UserID, uc.cfg.AccessTokenTTL); err != nil {
			logger.Error("revoke access tokens on reuse failed", "err", err)
		}
		return nil, ErrTokenRevoked
	}

	if stored.IsExpired() {
		return nil, ErrTokenExpired
	}

	u, err := uc.repo.GetUserByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}

	newRaw, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(uc.cfg.RefreshTokenTTL)
	if _, err := uc.repo.RotateRefreshToken(ctx, stored.ID, CreateRefreshTokenInput{
		UserID:    stored.UserID,
		TokenHash: auth.HashRefreshToken(newRaw),
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

// Logout revokes the supplied refresh token and adds the access-token jti to
// the denylist until its natural expiry.
func (uc *UseCase) Logout(ctx context.Context, refreshToken, accessJTI string, accessExpiresAt time.Time) error {
	if refreshToken != "" {
		hash := auth.HashRefreshToken(refreshToken)
		if stored, err := uc.repo.GetRefreshTokenByHash(ctx, hash); err == nil {
			if err := uc.repo.RevokeRefreshToken(ctx, stored.ID); err != nil {
				return err
			}
		} else if !errors.Is(err, ErrInvalidToken) {
			return err
		}
	}

	if accessJTI != "" {
		if err := uc.store.RevokeAccessJTI(ctx, accessJTI, accessExpiresAt); err != nil {
			return err
		}
	}
	return nil
}

// LogoutAll revokes every refresh token for the user and stamps a global
// access-token watermark in Redis so all currently-issued access tokens are
// rejected by the middleware.
func (uc *UseCase) LogoutAll(ctx context.Context, userID string) error {
	if err := uc.repo.RevokeAllUserRefreshTokens(ctx, userID); err != nil {
		return err
	}
	return uc.store.RevokeAllUserAccessTokens(ctx, userID, uc.cfg.AccessTokenTTL)
}

func (uc *UseCase) GetUser(ctx context.Context, id string) (*User, error) {
	return uc.repo.GetUserByID(ctx, id)
}

// VerifyAccessToken verifies the JWT signature/expiry, then checks the Redis
// denylist (jti) and the user's revoked-before watermark.
func (uc *UseCase) VerifyAccessToken(ctx context.Context, tokenStr string) (*auth.AccessClaims, error) {
	claims, err := uc.jwt.Verify(tokenStr)
	if err != nil {
		return nil, err
	}

	revoked, err := uc.store.IsAccessJTIRevoked(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, ErrUnauthorized
	}

	watermark, err := uc.store.RevokedBefore(ctx, claims.Subject)
	if err != nil {
		return nil, err
	}
	if watermark > 0 && claims.IssuedAt != nil && claims.IssuedAt.Unix() < watermark {
		return nil, ErrUnauthorized
	}
	return claims, nil
}

// --- helpers ---

func (uc *UseCase) issueTokenPair(ctx context.Context, userID string, meta SessionMeta) (*TokenPair, error) {
	access, err := uc.jwt.Issue(userID)
	if err != nil {
		return nil, err
	}

	rawRefresh, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(uc.cfg.RefreshTokenTTL)
	if _, err := uc.repo.CreateRefreshToken(ctx, CreateRefreshTokenInput{
		UserID:    userID,
		TokenHash: auth.HashRefreshToken(rawRefresh),
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
