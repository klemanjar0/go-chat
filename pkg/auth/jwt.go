package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type AccessClaims struct {
	jwt.RegisteredClaims
}

type JWTIssuer struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

func NewJWTIssuer(secret, issuer string, ttl time.Duration) *JWTIssuer {
	return &JWTIssuer{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
	}
}

type IssuedAccessToken struct {
	Token     string
	JTI       string
	ExpiresAt time.Time
	IssuedAt  time.Time
}

func (j *JWTIssuer) Issue(userID string) (*IssuedAccessToken, error) {
	now := time.Now().UTC()
	exp := now.Add(j.ttl)
	jti := uuid.NewString()

	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.secret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	return &IssuedAccessToken{
		Token:     signed,
		JTI:       jti,
		ExpiresAt: exp,
		IssuedAt:  now,
	}, nil
}

func (j *JWTIssuer) Verify(tokenStr string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	parsed, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.secret, nil
	}, jwt.WithIssuer(j.issuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	if !parsed.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
