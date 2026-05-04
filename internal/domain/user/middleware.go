package user

import (
	"strings"

	"go-chat/pkg/httputil"

	"github.com/gofiber/fiber/v3"
)

// AuthMiddleware verifies the bearer access token, checks Redis denylist /
// per-user revocation watermark, and stores the resolved user id, jti and
// expiry in the request locals for downstream handlers.
func AuthMiddleware(uc *UseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		if header == "" {
			return httputil.Respond(c).Unauthorized(ErrUnauthorized)
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			return httputil.Respond(c).Unauthorized(ErrUnauthorized)
		}
		token := strings.TrimSpace(header[len(prefix):])
		if token == "" {
			return httputil.Respond(c).Unauthorized(ErrUnauthorized)
		}

		claims, err := uc.VerifyAccessToken(c.Context(), token)
		if err != nil {
			return httputil.Respond(c).Unauthorized(ErrUnauthorized)
		}

		c.Locals(ContextUserIDKey, claims.Subject)
		c.Locals(ContextAccessJTIKey, claims.ID)
		if claims.ExpiresAt != nil {
			c.Locals(ContextAccessExpiresKey, claims.ExpiresAt.Time)
		}

		return c.Next()
	}
}
