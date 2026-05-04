package http

import (
	"go-chat/pkg/httputil"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(app *fiber.App, h UserHandler, authMiddleware fiber.Handler) {
	api := app.Group("/api")

	api.Get("/health", func(c fiber.Ctx) error {
		return httputil.Respond(c).OK(map[string]string{"status": "ok"})
	})

	// public auth
	api.Post("/auth/register", h.Register)
	api.Post("/auth/login", h.Login)
	api.Post("/auth/refresh", h.Refresh)

	// protected
	auth := api.Group("/auth", authMiddleware)
	auth.Post("/logout", h.Logout)
	auth.Post("/logout-all", h.LogoutAll)

	users := api.Group("/users", authMiddleware)
	users.Get("/me", h.Me)
}
