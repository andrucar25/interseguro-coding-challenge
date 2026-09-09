// Package handler provides the Fiber transport for QR factorization.
package handler

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// New constructs the HTTP application.
func New(factorizer factorizer, allowedOrigins string) *fiber.App {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: strings.Split(allowedOrigins, ","),
		AllowMethods: []string{fiber.MethodPost, fiber.MethodOptions},
		AllowHeaders: []string{fiber.HeaderContentType},
	}))
	app.Post("/qr", NewQRHandler(factorizer).Factorize)
	return app
}
