// Package httpapi provides the HTTP transport for QR factorization.
package httpapi

import (
	"context"

	"github.com/andrucar25/interseguro-coding-challenge/go-api/qr"
	"github.com/andrucar25/interseguro-coding-challenge/go-api/statistics"
	"github.com/gofiber/fiber/v3"
)

type statisticsCalculator interface {
	Calculate(context.Context, qr.Matrix, qr.Matrix) (statistics.Result, error)
}

// New constructs the HTTP application.
func New(statisticsClient statisticsCalculator) *fiber.App {
	app := fiber.New()
	app.Post("/qr", factorizeQR(statisticsClient))
	return app
}
