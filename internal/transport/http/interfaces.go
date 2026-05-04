package http

import "github.com/gofiber/fiber/v3"

type UserHandler interface {
	Register(c fiber.Ctx) error
	Login(c fiber.Ctx) error
	Refresh(c fiber.Ctx) error
	Logout(c fiber.Ctx) error
	LogoutAll(c fiber.Ctx) error
	Me(c fiber.Ctx) error
}
