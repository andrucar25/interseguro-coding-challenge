package httpapi

import (
	"errors"

	"github.com/andrucar25/interseguro-coding-challenge/go-api/qr"
	"github.com/andrucar25/interseguro-coding-challenge/go-api/statistics"
	"github.com/gofiber/fiber/v3"
)

type qrRequest struct {
	Matrix qr.Matrix `json:"matrix"`
}

type qrResponse struct {
	Q          qr.Matrix         `json:"q"`
	R          qr.Matrix         `json:"r"`
	Statistics statistics.Result `json:"statistics"`
}

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func factorizeQR(statisticsClient statisticsCalculator) fiber.Handler {
	return func(c fiber.Ctx) error {
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
		if statisticsClient == nil {
			return writeInternalError(c)
		}

		result, err := statisticsClient.Calculate(c.RequestCtx(), q, r)
		if err != nil {
			return writeStatisticsError(c, err)
		}

		return c.JSON(qrResponse{Q: q, R: r, Statistics: result})
	}
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

	return writeInternalError(c)
}

func writeStatisticsError(c fiber.Ctx, err error) error {
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
