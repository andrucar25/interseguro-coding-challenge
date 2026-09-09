// Package usecase coordinates QR factorization and statistics calculation.
package usecase

import (
	"context"
	"errors"

	"github.com/andrucar25/interseguro-coding-challenge/go-api/external/statistics"
	"github.com/andrucar25/interseguro-coding-challenge/go-api/qr"
)

var errNilStatisticsCalculator = errors.New("statistics calculator must not be nil")

type statisticsCalculator interface {
	Calculate(context.Context, qr.Matrix, qr.Matrix) (statistics.Result, error)
}

// Result contains the QR factorization and its downstream statistics.
type Result struct {
	Q          qr.Matrix
	R          qr.Matrix
	Statistics statistics.Result
}

// Service coordinates QR factorization with the statistics collaborator.
type Service struct {
	statisticsCalculator statisticsCalculator
}

// New constructs a QR factorization service.
func New(statisticsCalculator statisticsCalculator) (*Service, error) {
	if statisticsCalculator == nil {
		return nil, errNilStatisticsCalculator
	}

	return &Service{statisticsCalculator: statisticsCalculator}, nil
}

// Factorize produces the QR matrices and obtains their statistics.
func (s *Service) Factorize(ctx context.Context, matrix qr.Matrix) (Result, error) {
	q, r, err := qr.Factorize(matrix)
	if err != nil {
		return Result{}, err
	}

	statisticsResult, err := s.statisticsCalculator.Calculate(ctx, q, r)
	if err != nil {
		return Result{}, err
	}

	return Result{Q: q, R: r, Statistics: statisticsResult}, nil
}
