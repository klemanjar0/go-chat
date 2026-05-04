package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// AccessTokenStore handles ephemeral access-token state in Redis:
//   - per-jti denylist for revoked access tokens (logout)
//   - per-user "revoked-before" watermark to invalidate every access token
//     issued before a given time (logout-all / password change)
type AccessTokenStore struct {
	rdb *redis.Client
}

func NewAccessTokenStore(rdb *redis.Client) *AccessTokenStore {
	return &AccessTokenStore{rdb: rdb}
}

func jtiKey(jti string) string               { return "auth:revoked:jti:" + jti }
func userRevokedBeforeKey(uid string) string { return "auth:revoked-before:user:" + uid }

func (s *AccessTokenStore) RevokeJTI(ctx context.Context, jti string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return s.rdb.Set(ctx, jtiKey(jti), "1", ttl).Err()
}

func (s *AccessTokenStore) IsJTIRevoked(ctx context.Context, jti string) (bool, error) {
	n, err := s.rdb.Exists(ctx, jtiKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// RevokeAllForUser stamps a revoked-before timestamp for the user. Any access
// token whose iat is older than this stamp must be rejected. The key expires
// after the longest possible access-token lifetime so it self-cleans.
func (s *AccessTokenStore) RevokeAllForUser(ctx context.Context, userID string, accessTTL time.Duration) error {
	now := time.Now().Unix()
	return s.rdb.Set(ctx, userRevokedBeforeKey(userID), strconv.FormatInt(now, 10), accessTTL+time.Minute).Err()
}

// RevokedBefore returns the unix-second watermark for the user, or 0 if none.
func (s *AccessTokenStore) RevokedBefore(ctx context.Context, userID string) (int64, error) {
	val, err := s.rdb.Get(ctx, userRevokedBeforeKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
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
