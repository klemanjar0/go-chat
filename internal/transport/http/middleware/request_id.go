package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// RequestID propagates the X-Request-ID header (or generates one) and stores
// it in the request locals so handlers and the logger can correlate calls.
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Get(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(RequestIDHeader, id)
		c.Locals(RequestIDKey, id)
		return c.Next()
	}
}
