// Package httpapi provides the HTTP transport for QR factorization.
package httpapi

import "github.com/gofiber/fiber/v3"

// New constructs the HTTP application.
func New() *fiber.App {
	app := fiber.New()
	app.Post("/qr", factorizeQR)
	return app
}
