package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-chat/internal/config"
	domainauth "go-chat/internal/domain/auth"
	"go-chat/internal/domain/user"
	authuc "go-chat/internal/usecase/auth"
	pkgauth "go-chat/pkg/auth"
)

// --- in-memory fakes ---

type fakeUserStore struct {
	byID map[string]*user.User
}

func (f *fakeUserStore) Create(_ context.Context, _ user.CreateInput) (*user.User, error) {
	panic("not used")
}
func (f *fakeUserStore) GetByID(_ context.Context, id string) (*user.User, error) {
	if u, ok := f.byID[id]; ok {
		return u, nil
	}
	return nil, user.ErrUserNotFound
}
func (f *fakeUserStore) GetByUsername(_ context.Context, _ string) (*user.User, error) {
	panic("not used")
}

type fakeRefreshStore struct {
	byHash               map[string]*domainauth.RefreshToken
	revokeAllForUserHits int
}

func (f *fakeRefreshStore) Create(_ context.Context, _ domainauth.CreateRefreshTokenInput) (*domainauth.RefreshToken, error) {
	panic("not used")
}
func (f *fakeRefreshStore) GetByHash(_ context.Context, hash string) (*domainauth.RefreshToken, error) {
	if rt, ok := f.byHash[hash]; ok {
		return rt, nil
	}
	return nil, domainauth.ErrInvalidToken
}
func (f *fakeRefreshStore) Revoke(_ context.Context, _ string) error { return nil }
func (f *fakeRefreshStore) RevokeAllForUser(_ context.Context, _ string) error {
	f.revokeAllForUserHits++
	return nil
}
func (f *fakeRefreshStore) Rotate(_ context.Context, _ string, _ domainauth.CreateRefreshTokenInput) (*domainauth.RefreshToken, error) {
	panic("not used")
}

type fakeAccessStore struct {
	revokeAllForUserHits int
}

func (f *fakeAccessStore) RevokeJTI(_ context.Context, _ string, _ time.Time) error { return nil }
func (f *fakeAccessStore) IsJTIRevoked(_ context.Context, _ string) (bool, error)   { return false, nil }
func (f *fakeAccessStore) RevokeAllForUser(_ context.Context, _ string, _ time.Duration) error {
	f.revokeAllForUserHits++
	return nil
}
func (f *fakeAccessStore) RevokedBefore(_ context.Context, _ string) (int64, error) {
	return 0, nil
}

// TestRefresh_TokenReuseRevokesEverything verifies the reuse-detection path:
// presenting an already-revoked refresh token must revoke every token for
// that user (refresh + access) and surface ErrTokenRevoked.
func TestRefresh_TokenReuseRevokesEverything(t *testing.T) {
	t.Parallel()

	const raw = "any-refresh-token"
	hash := pkgauth.HashRefreshToken(raw)
	revokedAt := time.Now().Add(-1 * time.Minute)

	users := &fakeUserStore{byID: map[string]*user.User{}}
	refresh := &fakeRefreshStore{byHash: map[string]*domainauth.RefreshToken{
		hash: {
			ID:        "tok-1",
			UserID:    "user-1",
			TokenHash: hash,
			ExpiresAt: time.Now().Add(time.Hour),
			RevokedAt: &revokedAt,
		},
	}}
	access := &fakeAccessStore{}

	jwt := pkgauth.NewJWTIssuer("test-secret", "test", time.Minute)
	uc := authuc.NewUseCase(config.Auth{
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	}, users, refresh, access, jwt)

	_, err := uc.Refresh(context.Background(), raw, authuc.SessionMeta{})
	if !errors.Is(err, domainauth.ErrTokenRevoked) {
		t.Fatalf("expected ErrTokenRevoked, got %v", err)
	}
	if refresh.revokeAllForUserHits != 1 {
		t.Errorf("refresh.RevokeAllForUser hits = %d, want 1", refresh.revokeAllForUserHits)
	}
	if access.revokeAllForUserHits != 1 {
		t.Errorf("access.RevokeAllForUser hits = %d, want 1", access.revokeAllForUserHits)
	}
}

// TestRefresh_ExpiredToken verifies expired (but not revoked) tokens return
// ErrTokenExpired without triggering the reuse-revocation cascade.
func TestRefresh_ExpiredToken(t *testing.T) {
	t.Parallel()

	const raw = "any-refresh-token"
	hash := pkgauth.HashRefreshToken(raw)

	users := &fakeUserStore{byID: map[string]*user.User{}}
	refresh := &fakeRefreshStore{byHash: map[string]*domainauth.RefreshToken{
		hash: {
			ID:        "tok-1",
			UserID:    "user-1",
			TokenHash: hash,
			ExpiresAt: time.Now().Add(-time.Minute),
		},
	}}
	access := &fakeAccessStore{}

	jwt := pkgauth.NewJWTIssuer("test-secret", "test", time.Minute)
	uc := authuc.NewUseCase(config.Auth{AccessTTL: time.Minute, RefreshTTL: time.Hour},
		users, refresh, access, jwt)

	_, err := uc.Refresh(context.Background(), raw, authuc.SessionMeta{})
	if !errors.Is(err, domainauth.ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
	if refresh.revokeAllForUserHits != 0 || access.revokeAllForUserHits != 0 {
		t.Errorf("unexpected revoke-all calls on plain expiry: refresh=%d access=%d",
			refresh.revokeAllForUserHits, access.revokeAllForUserHits)
	}
}
