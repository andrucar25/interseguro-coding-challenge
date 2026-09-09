package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/andrucar25/interseguro-coding-challenge/go-api/external/statistics"
	"github.com/andrucar25/interseguro-coding-challenge/go-api/qr"
)

func TestFactorizeReturnsMatricesAndStatistics(t *testing.T) {
	calculator := &fakeStatisticsCalculator{result: statistics.Result{Maximum: 4, Minimum: -1, Sum: 5, Average: 1.25, HasDiagonalMatrix: true}}
	service := newService(t, calculator)
	input := qr.Matrix{{3}, {4}}

	got, err := service.Factorize(context.Background(), input)
	if err != nil {
		t.Fatalf("Factorize() error = %v", err)
	}
	if !matricesEqual(got.Q, calculator.q) || !matricesEqual(got.R, calculator.r) {
		t.Errorf("Calculate() matrices = q=%v r=%v, want q=%v r=%v", calculator.q, calculator.r, got.Q, got.R)
	}
	if got.Statistics != calculator.result {
		t.Errorf("Statistics = %#v, want %#v", got.Statistics, calculator.result)
	}
}

func TestFactorizeDoesNotCallStatisticsWhenQRFails(t *testing.T) {
	calculator := &fakeStatisticsCalculator{}
	service := newService(t, calculator)

	_, err := service.Factorize(context.Background(), qr.Matrix{})
	if !errors.Is(err, qr.ErrEmptyMatrix) {
		t.Fatalf("Factorize() error = %v, want errors.Is(..., %v)", err, qr.ErrEmptyMatrix)
	}
	if calculator.called {
		t.Error("Calculate() was called after QR factorization failed")
	}
}

func TestFactorizePropagatesStatisticsError(t *testing.T) {
	want := errors.New("statistics unavailable")
	calculator := &fakeStatisticsCalculator{err: want}
	service := newService(t, calculator)

	_, err := service.Factorize(context.Background(), qr.Matrix{{1}})
	if !errors.Is(err, want) {
		t.Fatalf("Factorize() error = %v, want errors.Is(..., %v)", err, want)
	}
}

func TestNewRejectsNilStatisticsCalculator(t *testing.T) {
	if _, err := New(nil); !errors.Is(err, errNilStatisticsCalculator) {
		t.Fatalf("New(nil) error = %v, want errors.Is(..., %v)", err, errNilStatisticsCalculator)
	}
}

func newService(t *testing.T, calculator statisticsCalculator) *Service {
	t.Helper()
	service, err := New(calculator)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return service
}

type fakeStatisticsCalculator struct {
	result statistics.Result
	err    error
	called bool
	q      qr.Matrix
	r      qr.Matrix
}

func (calculator *fakeStatisticsCalculator) Calculate(_ context.Context, q, r qr.Matrix) (statistics.Result, error) {
	calculator.called = true
	calculator.q = q
	calculator.r = r
	return calculator.result, calculator.err
}

func matricesEqual(left, right qr.Matrix) bool {
	if len(left) != len(right) {
		return false
	}
	for row := range left {
		if len(left[row]) != len(right[row]) {
			return false
		}
		for column := range left[row] {
			if left[row][column] != right[row][column] {
				return false
			}
		}
	}
	return true
}
