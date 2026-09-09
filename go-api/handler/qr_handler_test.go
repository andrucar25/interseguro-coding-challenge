package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrucar25/interseguro-coding-challenge/go-api/external/statistics"
	"github.com/andrucar25/interseguro-coding-challenge/go-api/qr"
	"github.com/andrucar25/interseguro-coding-challenge/go-api/usecase"
	"github.com/gofiber/fiber/v3"
)

const allowedOrigin = "http://localhost:5173"

func TestFactorizeQRReturnsOperationResponse(t *testing.T) {
	want := usecase.Result{
		Q:          qr.Matrix{{0.6}, {0.8}},
		R:          qr.Matrix{{5}},
		Statistics: statistics.Result{Maximum: 5, Minimum: 0.6, Sum: 6.4, Average: 2.1333333333333333, HasDiagonalMatrix: false},
	}
	operation := &fakeFactorizer{result: want}
	response := performRequest(t, New(operation, allowedOrigin), []byte(`{"matrix":[[3],[4]]}`))

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	var got qrResponse
	decodeJSON(t, response.Body, &got)
	if !matricesEqual(got.Q, want.Q) || !matricesEqual(got.R, want.R) || got.Statistics != want.Statistics {
		t.Errorf("response = %#v, want %#v", got, want)
	}
	if !matricesEqual(operation.matrix, qr.Matrix{{3}, {4}}) {
		t.Errorf("operation matrix = %v, want %v", operation.matrix, qr.Matrix{{3}, {4}})
	}
}

func TestFactorizeQRRejectsInvalidJSON(t *testing.T) {
	operation := &fakeFactorizer{}
	response := performRequest(t, New(operation, allowedOrigin), []byte(`{"matrix":`))

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
	var got errorResponse
	decodeJSON(t, response.Body, &got)
	want := errorResponse{Error: "invalid_request", Message: "request body must contain a valid matrix"}
	if got != want {
		t.Errorf("response = %#v, want %#v", got, want)
	}
	if operation.called {
		t.Error("operation was called for malformed JSON")
	}
}

func TestFactorizeQRMapsErrors(t *testing.T) {
	testCases := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   errorResponse
	}{
		{
			name:       "invalid matrix",
			err:        qr.ErrNonRectangular,
			wantStatus: http.StatusBadRequest,
			wantBody:   errorResponse{Error: "invalid_request", Message: qr.ErrNonRectangular.Error()},
		},
		{
			name:       "unavailable downstream",
			err:        wrapError(statistics.ErrDownstream),
			wantStatus: http.StatusBadGateway,
			wantBody:   errorResponse{Error: "downstream_error", Message: "unable to obtain statistics"},
		},
		{
			name:       "timeout",
			err:        wrapError(statistics.ErrTimeout),
			wantStatus: http.StatusGatewayTimeout,
			wantBody:   errorResponse{Error: "downstream_timeout", Message: "statistics service timed out"},
		},
		{
			name:       "unexpected error",
			err:        errors.New("unexpected failure"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   errorResponse{Error: "internal_error", Message: "unable to factorize matrix"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			response := performRequest(t, New(&fakeFactorizer{err: testCase.err}, allowedOrigin), []byte(`{"matrix":[[1]]}`))
			if response.StatusCode != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", response.StatusCode, testCase.wantStatus)
			}
			var got errorResponse
			decodeJSON(t, response.Body, &got)
			if got != testCase.wantBody {
				t.Errorf("response = %#v, want %#v", got, testCase.wantBody)
			}
		})
	}
}

func TestFactorizeQRNilOperationReturnsInternalError(t *testing.T) {
	response := performRequest(t, New(nil, allowedOrigin), []byte(`{"matrix":[[1]]}`))
	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusInternalServerError)
	}
}

func TestCORS(t *testing.T) {
	app := New(&fakeFactorizer{}, allowedOrigin)

	preflightRequest := httptest.NewRequest(http.MethodOptions, "/qr", nil)
	preflightRequest.Header.Set("Origin", allowedOrigin)
	preflightRequest.Header.Set("Access-Control-Request-Method", http.MethodPost)
	preflightRequest.Header.Set("Access-Control-Request-Headers", "Content-Type")
	preflightResponse, err := app.Test(preflightRequest)
	if err != nil {
		t.Fatalf("app.Test() preflight error = %v", err)
	}
	defer preflightResponse.Body.Close()
	if preflightResponse.StatusCode != http.StatusNoContent {
		t.Errorf("preflight status = %d, want %d", preflightResponse.StatusCode, http.StatusNoContent)
	}
	if got := preflightResponse.Header.Get("Access-Control-Allow-Origin"); got != allowedOrigin {
		t.Errorf("preflight Access-Control-Allow-Origin = %q, want %q", got, allowedOrigin)
	}
	if got := preflightResponse.Header.Get("Access-Control-Allow-Methods"); got != "POST, OPTIONS" {
		t.Errorf("preflight Access-Control-Allow-Methods = %q, want %q", got, "POST, OPTIONS")
	}
	if got := preflightResponse.Header.Get("Access-Control-Allow-Headers"); got != "Content-Type" {
		t.Errorf("preflight Access-Control-Allow-Headers = %q, want %q", got, "Content-Type")
	}

	for _, origin := range []string{allowedOrigin, "http://untrusted.example"} {
		t.Run(origin, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/qr", bytes.NewReader([]byte(`{"matrix":[[1]]}`)))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Origin", origin)

			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("app.Test() POST error = %v", err)
			}
			defer response.Body.Close()

			wantOrigin := ""
			if origin == allowedOrigin {
				wantOrigin = allowedOrigin
			}
			if got := response.Header.Get("Access-Control-Allow-Origin"); got != wantOrigin {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, wantOrigin)
			}
		})
	}
}

func performRequest(t *testing.T, app *fiber.App, body []byte) *http.Response {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/qr", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	return response
}

func decodeJSON(t *testing.T, body io.ReadCloser, destination any) {
	t.Helper()
	defer body.Close()
	if err := json.NewDecoder(body).Decode(destination); err != nil {
		t.Fatalf("decode response JSON: %v", err)
	}
}

func wrapError(cause error) error {
	return errors.Join(errors.New("wrapped failure"), cause)
}

type fakeFactorizer struct {
	result usecase.Result
	err    error
	called bool
	matrix qr.Matrix
}

func (factorizer *fakeFactorizer) Factorize(_ context.Context, matrix qr.Matrix) (usecase.Result, error) {
	factorizer.called = true
	factorizer.matrix = matrix
	return factorizer.result, factorizer.err
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
