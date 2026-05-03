package http

import (
	"go-chat/pkg/httputil"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(app *fiber.App, userHandler UserHandler, authMiddleware fiber.Handler) {
	api := app.Group("/api")

	api.Get("/health", func(c fiber.Ctx) error {
		return httputil.Respond(c).OK(map[string]string{})
	})

	// public
	// api.Post("/users", h.Register)
	// api.Post("/auth/login", h.Authenticate)
	// api.Post("/auth/refresh", h.RefreshToken)

	// // protected
	// auth := api.Group("/auth", authMiddleware)
	// auth.Post("/logout", h.Logout)

	// users := api.Group("/users", authMiddleware)
	// users.Get("/me", h.Me)
	// users.Get("/:id", h.GetUser)
	// users.Get("/email/:email", h.GetUserByEmail)
	// users.Get("/:id/validate", h.ValidateUser)
	// users.Post("/:id/change-password", h.ChangePassword)
}
