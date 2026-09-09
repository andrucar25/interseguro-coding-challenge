// Package httpapi provides the HTTP transport for QR factorization.
package httpapi

import (
	"context"
	"strings"

	"github.com/andrucar25/interseguro-coding-challenge/go-api/qr"
	"github.com/andrucar25/interseguro-coding-challenge/go-api/statistics"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

type statisticsCalculator interface {
	Calculate(context.Context, qr.Matrix, qr.Matrix) (statistics.Result, error)
}

// New constructs the HTTP application.
func New(statisticsClient statisticsCalculator, allowedOrigins string) *fiber.App {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: strings.Split(allowedOrigins, ","),
		AllowMethods: []string{fiber.MethodPost, fiber.MethodOptions},
		AllowHeaders: []string{fiber.HeaderContentType},
	}))
	app.Post("/qr", factorizeQR(statisticsClient))
	return app
}
