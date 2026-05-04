package auth

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

type CreateRefreshTokenInput struct {
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	UserAgent *string
	IP        *netip.Addr
}

func (rt *RefreshToken) IsRevoked() bool             { return rt.RevokedAt != nil }
func (rt *RefreshToken) IsExpired(now time.Time) bool { return now.After(rt.ExpiresAt) }

func refreshTokenFromPg(row *sqlcgen.RefreshToken) *RefreshToken {
	rt := &RefreshToken{
		ID:        utilid.FromPg(row.ID).AsString(),
		UserID:    utilid.FromPg(row.UserID).AsString(),
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt.Time,
		UserAgent: row.UserAgent,
		IP:        row.Ip,
		CreatedAt: row.CreatedAt.Time,
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		rt.RevokedAt = &t
	}
	if row.ReplacedByID.Valid {
		v := utilid.FromPg(row.ReplacedByID).AsString()
		rt.ReplacedByID = &v
	}
	return rt
}
