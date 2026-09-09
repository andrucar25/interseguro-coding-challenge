package httpapi

import (
	"errors"

	"github.com/andrucar25/interseguro-coding-challenge/go-api/qr"
	"github.com/gofiber/fiber/v3"
)

type qrRequest struct {
	Matrix qr.Matrix `json:"matrix"`
}

type qrResponse struct {
	Q qr.Matrix `json:"q"`
	R qr.Matrix `json:"r"`
}

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func factorizeQR(c fiber.Ctx) error {
	var request qrRequest
	if err := c.Bind().JSON(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse{
			Error:   "invalid_request",
			Message: "request body must contain a valid matrix",
		})
	}

	q, r, err := qr.Factorize(request.Matrix)
	if err != nil {
		return writeFactorizationError(c, err)
	}

	return c.JSON(qrResponse{Q: q, R: r})
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

	return c.Status(fiber.StatusInternalServerError).JSON(errorResponse{
		Error:   "internal_error",
		Message: "unable to factorize matrix",
	})
}
