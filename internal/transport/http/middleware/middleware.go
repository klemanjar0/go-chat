// Package middleware holds cross-cutting HTTP middleware. Keys exposed here
// are written into fiber.Ctx locals by middleware and read by handlers.
package middleware

type ctxKey string

const (
	UserIDKey         ctxKey = "auth.user_id"
	AccessJTIKey      ctxKey = "auth.jti"
	AccessExpiresKey  ctxKey = "auth.exp"
	RequestIDKey      ctxKey = "request.id"
	RequestIDHeader          = "X-Request-ID"
)
