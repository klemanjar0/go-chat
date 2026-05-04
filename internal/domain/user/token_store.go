package user

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenStore handles ephemeral auth state in Redis:
//   - per-jti denylist for revoked access tokens (logout)
//   - per-user "revoked-before" watermark used to invalidate every access
//     token issued before a given time (logout-all / password change)
type TokenStore struct {
	rdb *redis.Client
}

func NewTokenStore(rdb *redis.Client) *TokenStore {
	return &TokenStore{rdb: rdb}
}

func jtiKey(jti string) string             { return "auth:revoked:jti:" + jti }
func userRevokedBeforeKey(uid string) string { return "auth:revoked-before:user:" + uid }

func (s *TokenStore) RevokeAccessJTI(ctx context.Context, jti string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return s.rdb.Set(ctx, jtiKey(jti), "1", ttl).Err()
}

func (s *TokenStore) IsAccessJTIRevoked(ctx context.Context, jti string) (bool, error) {
	n, err := s.rdb.Exists(ctx, jtiKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// RevokeAllUserAccessTokens stamps a "revoked-before" timestamp for the user.
// Any access token whose iat is older than this stamp must be rejected.
// We expire the key after the longest possible access-token lifetime so it
// self-cleans when no longer load-bearing.
func (s *TokenStore) RevokeAllUserAccessTokens(ctx context.Context, userID string, accessTTL time.Duration) error {
	now := time.Now().Unix()
	return s.rdb.Set(ctx, userRevokedBeforeKey(userID), strconv.FormatInt(now, 10), accessTTL+time.Minute).Err()
}

// RevokedBefore returns the unix-second watermark for the user, or 0 if none.
func (s *TokenStore) RevokedBefore(ctx context.Context, userID string) (int64, error) {
	val, err := s.rdb.Get(ctx, userRevokedBeforeKey(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	ts, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse revoked-before: %w", err)
	}
	return ts, nil
}
