package auth

import (
	"context"
	"errors"
	"strings"

	domainauth "go-chat/internal/domain/auth"
	"go-chat/internal/domain/user"
	pkgauth "go-chat/pkg/auth"
)

func (uc *UseCase) Login(ctx context.Context, username, password string, meta SessionMeta) (*AuthResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, domainauth.ErrInvalidCredentials
	}

	u, err := uc.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, domainauth.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := pkgauth.VerifyPassword(u.PasswordHash, password); err != nil {
		return nil, domainauth.ErrInvalidCredentials
	}

	tokens, err := uc.issueTokenPair(ctx, u.ID, meta)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: u, Tokens: *tokens}, nil
}
