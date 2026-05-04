package http

import (
	authhttp "go-chat/internal/transport/http/handlers/auth"
	userhttp "go-chat/internal/transport/http/handlers/user"
	"go-chat/pkg/httputil"

	"github.com/gofiber/fiber/v3"
)

type Handlers struct {
	Auth *authhttp.Handler
	User *userhttp.Handler
}

func RegisterRoutes(app *fiber.App, h Handlers, authMW fiber.Handler) {
	api := app.Group("/api")

	api.Get("/health", func(c fiber.Ctx) error {
		return httputil.Respond(c).OK(map[string]string{"status": "ok"})
	})

	// public auth
	api.Post("/auth/register", h.Auth.Register)
	api.Post("/auth/login", h.Auth.Login)
	api.Post("/auth/refresh", h.Auth.Refresh)

	// protected auth
	auth := api.Group("/auth", authMW)
	auth.Post("/logout", h.Auth.Logout)
	auth.Post("/logout-all", h.Auth.LogoutAll)

	// protected users
	users := api.Group("/users", authMW)
	users.Get("/me", h.User.Me)
}
