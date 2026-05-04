package user

import (
	"net/netip"
	"time"

	"go-chat/internal/db/sqlcgen"
	"go-chat/pkg/utilid"
)

type RefreshToken struct {
	ID           string
	UserID       string
	TokenHash    string
	ExpiresAt    time.Time
	RevokedAt    *time.Time
	ReplacedByID *string
	UserAgent    *string
	IP           *netip.Addr
	CreatedAt    time.Time
}

func RefreshTokenFromPg(t *sqlcgen.RefreshToken) *RefreshToken {
	rt := &RefreshToken{
		ID:        utilid.FromPg(t.ID).AsString(),
		UserID:    utilid.FromPg(t.UserID).AsString(),
		TokenHash: t.TokenHash,
		ExpiresAt: t.ExpiresAt.Time,
		UserAgent: t.UserAgent,
		IP:        t.Ip,
		CreatedAt: t.CreatedAt.Time,
	}
	if t.RevokedAt.Valid {
		v := t.RevokedAt.Time
		rt.RevokedAt = &v
	}
	if t.ReplacedByID.Valid {
		v := utilid.FromPg(t.ReplacedByID).AsString()
		rt.ReplacedByID = &v
	}
	return rt
}

func (rt *RefreshToken) IsRevoked() bool { return rt.RevokedAt != nil }
func (rt *RefreshToken) IsExpired() bool { return time.Now().After(rt.ExpiresAt) }
