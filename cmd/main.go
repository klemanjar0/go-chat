package main

import (
	"go-chat/internal/app"
	"go-chat/pkg/logger"
)

func main() {
	if err := app.Run(); err != nil {
		logger.Fatal("app exited with error", "err", err)
	}
}
