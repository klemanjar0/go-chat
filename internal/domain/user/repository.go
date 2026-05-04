package user

import (
	"context"
	"errors"

	"go-chat/internal/db/sqlcgen"
	"go-chat/pkg/utilid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	queries *sqlcgen.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlcgen.New(pool)}
}

func (r *Repository) Create(ctx context.Context, in CreateInput) (*User, error) {
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
	return fromPg(&row), nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	uid := utilid.FromString(id)
	if !uid.Valid {
		return nil, ErrInvalidUserID
	}
	row, err := r.queries.GetUserByID(ctx, uid.AsPgUUID())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return fromPg(&row), nil
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (*User, error) {
	row, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return fromPg(&row), nil
}

func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	return r.queries.UsernameExists(ctx, username)
}
