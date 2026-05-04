package auth

import (
	"context"
	"errors"
	"strings"

	domainauth "go-chat/internal/domain/auth"
	"go-chat/internal/domain/user"
	pkgauth "go-chat/pkg/auth"

	"github.com/jackc/pgx/v5/pgconn"
)

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
		return nil, domainauth.ErrInvalidPayload
	}

	hash, err := pkgauth.HashPassword(in.Password, uc.cfg.BcryptCost)
	if err != nil {
		return nil, err
	}

	created, err := uc.users.Create(ctx, user.CreateInput{
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
			return nil, user.ErrUsernameTaken
		}
		return nil, err
	}

	tokens, err := uc.issueTokenPair(ctx, created.ID, meta)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: created, Tokens: *tokens}, nil
}
