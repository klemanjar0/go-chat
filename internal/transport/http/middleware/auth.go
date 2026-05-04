package middleware

import (
	"strings"

	domainauth "go-chat/internal/domain/auth"
	authuc "go-chat/internal/usecase/auth"
	"go-chat/pkg/httputil"

	"github.com/gofiber/fiber/v3"
)

// Auth verifies the bearer access token, checks the denylist / per-user
// revocation watermark, and stores the resolved user id, jti and expiry in
// the request locals for downstream handlers.
func Auth(uc *authuc.UseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		if header == "" {
			return httputil.Respond(c).Unauthorized(domainauth.ErrUnauthorized)
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			return httputil.Respond(c).Unauthorized(domainauth.ErrUnauthorized)
		}
		token := strings.TrimSpace(header[len(prefix):])
		if token == "" {
			return httputil.Respond(c).Unauthorized(domainauth.ErrUnauthorized)
		}

		claims, err := uc.VerifyAccessToken(c.Context(), token)
		if err != nil {
			return httputil.Respond(c).Unauthorized(domainauth.ErrUnauthorized)
		}

		c.Locals(UserIDKey, claims.Subject)
		c.Locals(AccessJTIKey, claims.ID)
		if claims.ExpiresAt != nil {
			c.Locals(AccessExpiresKey, claims.ExpiresAt.Time)
		}

		return c.Next()
	}
}
