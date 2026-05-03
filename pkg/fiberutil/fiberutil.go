package fiberutil

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

func NewApp() *fiber.App {
	return fiber.New(fiber.Config{
		BodyLimit:    4 * 1024 * 1024, // 4 MB
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})
}
