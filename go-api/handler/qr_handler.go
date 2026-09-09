package handler

import (
	"context"
	"errors"

	"github.com/andrucar25/interseguro-coding-challenge/go-api/external/statistics"
	"github.com/andrucar25/interseguro-coding-challenge/go-api/qr"
	"github.com/andrucar25/interseguro-coding-challenge/go-api/usecase"
	"github.com/gofiber/fiber/v3"
)

type factorizer interface {
	Factorize(context.Context, qr.Matrix) (usecase.Result, error)
}

// QRHandler handles client QR factorization requests.
type QRHandler struct {
	factorizer factorizer
}

// NewQRHandler constructs a QR request handler.
func NewQRHandler(factorizer factorizer) *QRHandler {
	return &QRHandler{factorizer: factorizer}
}

// Factorize decodes a matrix, runs the application operation, and writes its response.
func (h *QRHandler) Factorize(c fiber.Ctx) error {
	var request qrRequest
	if err := c.Bind().JSON(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse{
			Error:   "invalid_request",
			Message: "request body must contain a valid matrix",
		})
	}
	if h.factorizer == nil {
		return writeInternalError(c)
	}

	result, err := h.factorizer.Factorize(c.RequestCtx(), request.Matrix)
	if err != nil {
		return writeFactorizationError(c, err)
	}

	return c.JSON(qrResponse{Q: result.Q, R: result.R, Statistics: result.Statistics})
}

func writeFactorizationError(c fiber.Ctx, err error) error {
	if errors.Is(err, qr.ErrEmptyMatrix) ||
		errors.Is(err, qr.ErrEmptyRow) ||
		errors.Is(err, qr.ErrNonRectangular) ||
		errors.Is(err, qr.ErrWideMatrix) ||
		errors.Is(err, qr.ErrNonFinite) {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}
	if errors.Is(err, statistics.ErrTimeout) {
		return c.Status(fiber.StatusGatewayTimeout).JSON(errorResponse{
			Error:   "downstream_timeout",
			Message: "statistics service timed out",
		})
	}
	if errors.Is(err, statistics.ErrDownstream) {
		return c.Status(fiber.StatusBadGateway).JSON(errorResponse{
			Error:   "downstream_error",
			Message: "unable to obtain statistics",
		})
	}
	return writeInternalError(c)
}

func writeInternalError(c fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(errorResponse{
		Error:   "internal_error",
		Message: "unable to factorize matrix",
	})
}
