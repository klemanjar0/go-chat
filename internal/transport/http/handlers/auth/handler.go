// Package auth wires the auth use case to HTTP. DTOs and request decoding
// live in this package; the use case knows nothing about fiber.
package auth

import (
	"net/netip"
	"time"

	domainauth "go-chat/internal/domain/auth"
	"go-chat/internal/transport/http/middleware"
	authuc "go-chat/internal/usecase/auth"
	"go-chat/pkg/httputil"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	uc *authuc.UseCase
}

func NewHandler(uc *authuc.UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return httputil.Respond(c).BadRequest(domainauth.ErrInvalidPayload)
	}

	res, err := h.uc.Register(c.Context(), authuc.RegisterInput{
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		AvatarURL: req.AvatarURL,
	}, sessionMeta(c))
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).Created(toAuthResponse(res))
}

func (h *Handler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return httputil.Respond(c).BadRequest(domainauth.ErrInvalidPayload)
	}

	res, err := h.uc.Login(c.Context(), req.Username, req.Password, sessionMeta(c))
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).OK(toAuthResponse(res))
}

func (h *Handler) Refresh(c fiber.Ctx) error {
	var req RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return httputil.Respond(c).BadRequest(domainauth.ErrInvalidPayload)
	}

	res, err := h.uc.Refresh(c.Context(), req.RefreshToken, sessionMeta(c))
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).OK(toAuthResponse(res))
}

func (h *Handler) Logout(c fiber.Ctx) error {
	var req LogoutRequest
	// body is optional — refresh token may also be inferred from the access claim
	_ = c.Bind().Body(&req)

	jti, _ := c.Locals(middleware.AccessJTIKey).(string)
	exp, _ := c.Locals(middleware.AccessExpiresKey).(time.Time)

	if err := h.uc.Logout(c.Context(), req.RefreshToken, jti, exp); err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).NoContent()
}

func (h *Handler) LogoutAll(c fiber.Ctx) error {
	userID, ok := c.Locals(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		return httputil.Respond(c).Unauthorized(domainauth.ErrUnauthorized)
	}
	if err := h.uc.LogoutAll(c.Context(), userID); err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).NoContent()
}

func sessionMeta(c fiber.Ctx) authuc.SessionMeta {
	meta := authuc.SessionMeta{}
	if ua := c.Get(fiber.HeaderUserAgent); ua != "" {
		meta.UserAgent = &ua
	}
	if ipStr := c.IP(); ipStr != "" {
		if addr, err := netip.ParseAddr(ipStr); err == nil {
			meta.IP = &addr
		}
	}
	return meta
}
