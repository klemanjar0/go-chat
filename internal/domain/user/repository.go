package user

import (
	"context"
	"errors"
	"net/netip"
	"time"

	"go-chat/internal/db/sqlcgen"
	"go-chat/pkg/utilid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *sqlcgen.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:    pool,
		queries: sqlcgen.New(pool),
	}
}

// --- users ---

type CreateUserInput struct {
	Username     string
	PasswordHash string
	FirstName    string
	LastName     string
	AvatarURL    *string
	Phone        string
}

func (r *Repository) CreateUser(ctx context.Context, in CreateUserInput) (*User, error) {
	row, err := r.queries.CreateUser(ctx, sqlcgen.CreateUserParams{
		Username:     in.Username,
		PasswordHash: in.PasswordHash,
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		AvatarUrl:    in.AvatarURL,
		Phone:        in.Phone,
	})
	if err != nil {
		return nil, err
	}
	return FromPg(&row), nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	uid := utilid.FromString(id)
	if !uid.Valid {
		return nil, ErrUserNotFound
	}
	row, err := r.queries.GetUserByID(ctx, uid.AsPgUUID())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return FromPg(&row), nil
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	row, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return FromPg(&row), nil
}

func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	return r.queries.UsernameExists(ctx, username)
}

// --- refresh tokens ---

type CreateRefreshTokenInput struct {
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	UserAgent *string
	IP        *netip.Addr
}

func (r *Repository) CreateRefreshToken(ctx context.Context, in CreateRefreshTokenInput) (*RefreshToken, error) {
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
	return RefreshTokenFromPg(&row), nil
}

func (r *Repository) GetRefreshTokenByHash(ctx context.Context, hash string) (*RefreshToken, error) {
	row, err := r.queries.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	return RefreshTokenFromPg(&row), nil
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, id string) error {
	uid := utilid.FromString(id)
	if !uid.Valid {
		return ErrInvalidToken
	}
	_, err := r.queries.RevokeRefreshToken(ctx, uid.AsPgUUID())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	return nil
}

func (r *Repository) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	uid := utilid.FromString(userID)
	if !uid.Valid {
		return ErrInvalidPayload
	}
	_, err := r.queries.RevokeAllUserRefreshTokens(ctx, uid.AsPgUUID())
	return err
}

// RotateRefreshToken atomically revokes the old token, links it to the new
// token's ID, and inserts the new token. Performed in a single transaction.
func (r *Repository) RotateRefreshToken(
	ctx context.Context,
	oldID string,
	newToken CreateRefreshTokenInput,
) (*RefreshToken, error) {
	oldUID := utilid.FromString(oldID)
	newUserUID := utilid.FromString(newToken.UserID)
	if !oldUID.Valid || !newUserUID.Valid {
		return nil, ErrInvalidToken
	}

	var created sqlcgen.RefreshToken

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := r.queries.WithTx(tx)

	created, err = q.CreateRefreshToken(ctx, sqlcgen.CreateRefreshTokenParams{
		UserID:    newUserUID.AsPgUUID(),
		TokenHash: newToken.TokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: newToken.ExpiresAt, Valid: true},
		UserAgent: newToken.UserAgent,
		Ip:        newToken.IP,
	})
	if err != nil {
		return nil, err
	}

	if _, err = q.RotateRefreshToken(ctx, sqlcgen.RotateRefreshTokenParams{
		ID:           oldUID.AsPgUUID(),
		ReplacedByID: created.ID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTokenRevoked
		}
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return RefreshTokenFromPg(&created), nil
}
