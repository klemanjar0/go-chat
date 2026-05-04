// Package user wires the user use case to HTTP.
package user

import (
	domainauth "go-chat/internal/domain/auth"
	"go-chat/internal/transport/http/middleware"
	useruc "go-chat/internal/usecase/user"
	"go-chat/pkg/httputil"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	uc *useruc.UseCase
}

func NewHandler(uc *useruc.UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Me(c fiber.Ctx) error {
	userID, ok := c.Locals(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		return httputil.Respond(c).Unauthorized(domainauth.ErrUnauthorized)
	}
	u, err := h.uc.GetByID(c.Context(), userID)
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).OK(toResponse(u))
}
