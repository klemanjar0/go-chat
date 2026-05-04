package user

import (
	"net/netip"
	"time"

	"go-chat/pkg/httputil"

	"github.com/gofiber/fiber/v3"
)

type contextKey string

const (
	ContextUserIDKey         contextKey = "auth.user_id"
	ContextAccessJTIKey      contextKey = "auth.jti"
	ContextAccessExpiresKey  contextKey = "auth.exp"
)

type HttpHandler struct {
	uc *UseCase
}

func NewHttpHandler(uc *UseCase) *HttpHandler {
	return &HttpHandler{uc: uc}
}

func (h *HttpHandler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return httputil.Respond(c).BadRequest(ErrInvalidPayload)
	}

	res, err := h.uc.Register(c.Context(), RegisterInput{
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

	return httputil.Respond(c).Created(ToAuthResponse(res))
}

func (h *HttpHandler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return httputil.Respond(c).BadRequest(ErrInvalidPayload)
	}

	res, err := h.uc.Login(c.Context(), req.Username, req.Password, sessionMeta(c))
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).OK(ToAuthResponse(res))
}

func (h *HttpHandler) Refresh(c fiber.Ctx) error {
	var req RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return httputil.Respond(c).BadRequest(ErrInvalidPayload)
	}

	res, err := h.uc.Refresh(c.Context(), req.RefreshToken, sessionMeta(c))
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).OK(ToAuthResponse(res))
}

func (h *HttpHandler) Logout(c fiber.Ctx) error {
	var req LogoutRequest
	// body is optional — refresh token may also live in the auth header chain
	_ = c.Bind().Body(&req)

	jti, _ := c.Locals(ContextAccessJTIKey).(string)
	exp, _ := c.Locals(ContextAccessExpiresKey).(time.Time)

	if err := h.uc.Logout(c.Context(), req.RefreshToken, jti, exp); err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).NoContent()
}

func (h *HttpHandler) LogoutAll(c fiber.Ctx) error {
	userID, ok := c.Locals(ContextUserIDKey).(string)
	if !ok || userID == "" {
		return httputil.Respond(c).Unauthorized(ErrUnauthorized)
	}
	if err := h.uc.LogoutAll(c.Context(), userID); err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).NoContent()
}

func (h *HttpHandler) Me(c fiber.Ctx) error {
	userID, ok := c.Locals(ContextUserIDKey).(string)
	if !ok || userID == "" {
		return httputil.Respond(c).Unauthorized(ErrUnauthorized)
	}
	u, err := h.uc.GetUser(c.Context(), userID)
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).OK(ToUserResponse(u))
}

func sessionMeta(c fiber.Ctx) SessionMeta {
	meta := SessionMeta{}
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
