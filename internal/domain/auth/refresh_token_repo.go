package auth

import (
	"context"
	"errors"

	"go-chat/internal/db/sqlcgen"
	"go-chat/pkg/utilid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepository struct {
	pool    *pgxpool.Pool
	queries *sqlcgen.Queries
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool, queries: sqlcgen.New(pool)}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, in CreateRefreshTokenInput) (*RefreshToken, error) {
	uid := utilid.FromString(in.UserID)
	if !uid.Valid {
		return nil, ErrInvalidPayload
	}
	row, err := r.queries.CreateRefreshToken(ctx, sqlcgen.CreateRefreshTokenParams{
		UserID:    uid.AsPgUUID(),
		TokenHash: in.TokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: in.ExpiresAt, Valid: true},
		UserAgent: in.UserAgent,
		Ip:        in.IP,
	})
	if err != nil {
		return nil, err
	}
	return refreshTokenFromPg(&row), nil
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, hash string) (*RefreshToken, error) {
	row, err := r.queries.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	return refreshTokenFromPg(&row), nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	uid := utilid.FromString(id)
	if !uid.Valid {
		return ErrInvalidToken
	}
	_, err := r.queries.RevokeRefreshToken(ctx, uid.AsPgUUID())
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	return nil
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	uid := utilid.FromString(userID)
	if !uid.Valid {
		return ErrInvalidPayload
	}
	_, err := r.queries.RevokeAllUserRefreshTokens(ctx, uid.AsPgUUID())
	return err
}

// Rotate atomically marks oldID as revoked, links it to the new token's id,
// and inserts the new token. Returns ErrTokenRevoked if oldID was already
// revoked (token-reuse race).
func (r *RefreshTokenRepository) Rotate(
	ctx context.Context,
	oldID string,
	newToken CreateRefreshTokenInput,
) (*RefreshToken, error) {
	oldUID := utilid.FromString(oldID)
	newUserUID := utilid.FromString(newToken.UserID)
	if !oldUID.Valid || !newUserUID.Valid {
		return nil, ErrInvalidToken
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := r.queries.WithTx(tx)

	created, err := q.CreateRefreshToken(ctx, sqlcgen.CreateRefreshTokenParams{
		UserID:    newUserUID.AsPgUUID(),
		TokenHash: newToken.TokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: newToken.ExpiresAt, Valid: true},
		UserAgent: newToken.UserAgent,
		Ip:        newToken.IP,
	})
	if err != nil {
		return nil, err
	}

	if _, err := q.RotateRefreshToken(ctx, sqlcgen.RotateRefreshTokenParams{
		ID:           oldUID.AsPgUUID(),
		ReplacedByID: created.ID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTokenRevoked
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return refreshTokenFromPg(&created), nil
}
