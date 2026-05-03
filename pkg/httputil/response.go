package httputil

import (
	"errors"
	"go-chat/pkg/logger"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type ResponseWriter struct {
	c          fiber.Ctx
	statusCode int
	data       any
	err        error
	headers    map[string]string
	message    string
	sent       bool
}

func Respond(c fiber.Ctx) *ResponseWriter {
	return &ResponseWriter{
		c:          c,
		statusCode: http.StatusOK,
		headers:    make(map[string]string),
	}
}

func (rw *ResponseWriter) Status(code int) *ResponseWriter {
	if rw.sent {
		logger.Warn("attempted to set status after response was sent")
		return rw
	}
	rw.statusCode = code
	return rw
}

func (rw *ResponseWriter) Json(data any) *ResponseWriter {
	if rw.sent {
		logger.Warn("attempted to set JSON data after response was sent")
		return rw
	}
	rw.data = data
	return rw
}

func (rw *ResponseWriter) Error(err error) *ResponseWriter {
	if rw.sent {
		logger.Warn("attempted to set error after response was sent")
		return rw
	}
	rw.err = err
	return rw
}

func (rw *ResponseWriter) Message(msg string) *ResponseWriter {
	if rw.sent {
		logger.Warn("attempted to set message after response was sent")
		return rw
	}
	rw.message = msg
	return rw
}

func (rw *ResponseWriter) Header(key, value string) *ResponseWriter {
	if rw.sent {
		logger.Warn("attempted to set header after response was sent")
		return rw
	}
	rw.headers[key] = value
	return rw
}

// Send finalizes and sends the response.
func (rw *ResponseWriter) Send() error {
	if rw.sent {
		logger.Warn("attempted to send response multiple times")
		return nil
	}
	rw.sent = true

	for key, value := range rw.headers {
		rw.c.Set(key, value)
	}

	if rw.err != nil {
		return rw.sendErrorResponse()
	}

	return rw.sendSuccessResponse()
}

func (rw *ResponseWriter) sendErrorResponse() error {
	if rw.statusCode >= 200 && rw.statusCode < 300 {
		rw.statusCode = mapError(rw.err)
	}

	logger.Error("HTTP error response",
		"status", rw.statusCode,
		"error", rw.err,
		"message", rw.message,
	)

	return rw.c.Status(rw.statusCode).JSON(ErrorResponse{
		Error:   rw.err.Error(),
		Message: rw.message,
		Code:    rw.statusCode,
	})
}

func (rw *ResponseWriter) sendSuccessResponse() error {
	if rw.data != nil {
		return rw.c.Status(rw.statusCode).JSON(rw.data)
	}

	if rw.message != "" {
		return rw.c.Status(rw.statusCode).JSON(SuccessResponse{
			Success: true,
			Message: rw.message,
		})
	}

	return rw.c.SendStatus(rw.statusCode)
}

// --- shorthand methods ---

func (rw *ResponseWriter) OK(data any) error {
	return rw.Status(http.StatusOK).Json(data).Send()
}

func (rw *ResponseWriter) Created(data any) error {
	return rw.Status(http.StatusCreated).Json(data).Send()
}

func (rw *ResponseWriter) NoContent() error {
	rw.statusCode = http.StatusNoContent
	return rw.Send()
}

func (rw *ResponseWriter) BadRequest(err error) error {
	return rw.Status(http.StatusBadRequest).Error(err).Send()
}

func (rw *ResponseWriter) Unauthorized(err error) error {
	return rw.Status(http.StatusUnauthorized).Error(err).Send()
}

func (rw *ResponseWriter) Forbidden(err error) error {
	return rw.Status(http.StatusForbidden).Error(err).Send()
}

func (rw *ResponseWriter) NotFound(err error) error {
	return rw.Status(http.StatusNotFound).Error(err).Send()
}

func (rw *ResponseWriter) Conflict(err error) error {
	return rw.Status(http.StatusConflict).Error(err).Send()
}

func (rw *ResponseWriter) InternalError(err error) error {
	return rw.Status(http.StatusInternalServerError).Error(err).Send()
}

// --- error mapping registry ---

type errorMapping struct {
	target error
	code   int
}

var mappings []errorMapping

// RegisterErrorMapping maps a sentinel error to an HTTP status code.
// Call at service startup to register domain errors.
//
//	httputil.RegisterErrorMapping(domain.ErrUserNotFound, 404)
//	httputil.RegisterErrorMapping(domain.ErrInvalidCredentials, 401)
func RegisterErrorMapping(target error, code int) {
	mappings = append(mappings, errorMapping{target: target, code: code})
}

func mapError(err error) int {
	for _, m := range mappings {
		if errors.Is(err, m.target) {
			return m.code
		}
	}
	return http.StatusInternalServerError
}
