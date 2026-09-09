package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrucar25/interseguro-coding-challenge/go-api/qr"
	"github.com/andrucar25/interseguro-coding-challenge/go-api/statistics"
	"github.com/gofiber/fiber/v3"
)

const testTolerance = 1e-10
const allowedOrigin = "http://localhost:5173"

func TestFactorizeQR(t *testing.T) {
	input := qr.Matrix{
		{12, -51, 4},
		{6, 167, -68},
		{-4, 24, -41},
	}
	body, err := json.Marshal(qrRequest{Matrix: input})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	response := performRequest(t, New(fakeStatisticsClient{result: statistics.Result{Maximum: 1, Minimum: -1, Sum: 0, Average: 0, HasDiagonalMatrix: false}}, allowedOrigin), body)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}

	var result qrResponse
	decodeJSON(t, response.Body, &result)
	assertShape(t, result.Q, 3, 3)
	assertShape(t, result.R, 3, 3)
	assertMatrixApproxEqual(t, multiply(transpose(result.Q), result.Q), identity(3))
	assertMatrixApproxEqual(t, multiply(result.Q, result.R), input)
	assertUpperTriangular(t, result.R)
}

func TestFactorizeQRCallsNodeAndReturnsStatistics(t *testing.T) {
	input := qr.Matrix{{3}, {4}}
	wantStatistics := statistics.Result{Maximum: 4, Minimum: -5, Sum: -1, Average: -0.25, HasDiagonalMatrix: true}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %q, want %q", request.Method, http.MethodPost)
		}
		if request.URL.Path != "/api/v1/statistics" {
			t.Errorf("path = %q, want %q", request.URL.Path, "/api/v1/statistics")
		}

		var payload struct {
			Q qr.Matrix `json:"q"`
			R qr.Matrix `json:"r"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode Node request: %v", err)
		}
		assertShape(t, payload.Q, 2, 1)
		assertShape(t, payload.R, 1, 1)
		assertMatrixApproxEqual(t, multiply(payload.Q, payload.R), input)

		if err := json.NewEncoder(writer).Encode(wantStatistics); err != nil {
			t.Errorf("encode Node response: %v", err)
		}
	}))
	defer server.Close()

	statisticsClient, err := statistics.New(server.URL, server.Client())
	if err != nil {
		t.Fatalf("statistics.New() error = %v", err)
	}
	body, err := json.Marshal(qrRequest{Matrix: input})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	response := performRequest(t, New(statisticsClient, allowedOrigin), body)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}

	var result qrResponse
	decodeJSON(t, response.Body, &result)
	assertMatrixApproxEqual(t, multiply(result.Q, result.R), input)
	if result.Statistics != wantStatistics {
		t.Errorf("statistics = %#v, want %#v", result.Statistics, wantStatistics)
	}
}

func TestFactorizeQRInvalidRequest(t *testing.T) {
	testCases := []struct {
		name    string
		body    string
		message string
	}{
		{name: "malformed JSON", body: `{"matrix":`, message: "request body must contain a valid matrix"},
		{name: "missing matrix", body: `{}`, message: qr.ErrEmptyMatrix.Error()},
		{name: "empty matrix", body: `{"matrix":[]}`, message: qr.ErrEmptyMatrix.Error()},
		{name: "empty row", body: `{"matrix":[[]]}`, message: qr.ErrEmptyRow.Error()},
		{name: "ragged rows", body: `{"matrix":[[1,2],[3]]}`, message: qr.ErrNonRectangular.Error()},
		{name: "wide matrix", body: `{"matrix":[[1,2,3],[4,5,6]]}`, message: qr.ErrWideMatrix.Error()},
		{name: "non-numeric value", body: `{"matrix":[["not-a-number"]]}`, message: "request body must contain a valid matrix"},
		{name: "non-finite value", body: `{"matrix":[[1e9999]]}`, message: "request body must contain a valid matrix"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			response := performRequest(t, New(fakeStatisticsClient{}, allowedOrigin), []byte(testCase.body))
			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusBadRequest)
			}

			var result errorResponse
			decodeJSON(t, response.Body, &result)
			if result.Error != "invalid_request" {
				t.Errorf("error = %q, want %q", result.Error, "invalid_request")
			}
			if result.Message != testCase.message {
				t.Errorf("message = %q, want %q", result.Message, testCase.message)
			}
		})
	}
}

func TestFactorizeQRMapsDownstreamErrors(t *testing.T) {
	testCases := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   errorResponse
	}{
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
			name:       "malformed downstream JSON",
			err:        wrapError(statistics.ErrDownstream),
			wantStatus: http.StatusBadGateway,
			wantBody:   errorResponse{Error: "downstream_error", Message: "unable to obtain statistics"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			response := performRequest(t, New(fakeStatisticsClient{err: testCase.err}, allowedOrigin), []byte(`{"matrix":[[1]]}`))
			if response.StatusCode != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", response.StatusCode, testCase.wantStatus)
			}
			var result errorResponse
			decodeJSON(t, response.Body, &result)
			if result != testCase.wantBody {
				t.Errorf("response = %#v, want %#v", result, testCase.wantBody)
			}
		})
	}
}

func TestCORS(t *testing.T) {
	app := New(fakeStatisticsClient{}, allowedOrigin)

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

type fakeStatisticsClient struct {
	result statistics.Result
	err    error
}

func (client fakeStatisticsClient) Calculate(context.Context, qr.Matrix, qr.Matrix) (statistics.Result, error) {
	return client.result, client.err
}

func wrapError(cause error) error {
	return errors.Join(errors.New("wrapped downstream failure"), cause)
}

func decodeJSON(t *testing.T, body io.ReadCloser, destination any) {
	t.Helper()
	defer body.Close()
	if err := json.NewDecoder(body).Decode(destination); err != nil {
		t.Fatalf("decode response JSON: %v", err)
	}
}

func assertShape(t *testing.T, matrix qr.Matrix, rows, columns int) {
	t.Helper()
	if len(matrix) != rows {
		t.Fatalf("matrix rows = %d, want %d", len(matrix), rows)
	}
	for _, row := range matrix {
		if len(row) != columns {
			t.Fatalf("matrix columns = %d, want %d", len(row), columns)
		}
	}
}

func assertMatrixApproxEqual(t *testing.T, got, want qr.Matrix) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("matrix row count = %d, want %d", len(got), len(want))
	}
	for row := range got {
		if len(got[row]) != len(want[row]) {
			t.Fatalf("matrix column count at row %d = %d, want %d", row, len(got[row]), len(want[row]))
		}
		for column := range got[row] {
			if !approximatelyEqual(got[row][column], want[row][column]) {
				t.Fatalf("matrix[%d][%d] = %.16g, want %.16g", row, column, got[row][column], want[row][column])
			}
		}
	}
}

func assertUpperTriangular(t *testing.T, matrix qr.Matrix) {
	t.Helper()
	for row := range matrix {
		for column := 0; column < row; column++ {
			if !approximatelyEqual(matrix[row][column], 0) {
				t.Fatalf("matrix[%d][%d] = %.16g, want approximately zero", row, column, matrix[row][column])
			}
		}
	}
}

func approximatelyEqual(got, want float64) bool {
	return math.Abs(got-want) <= testTolerance*math.Max(1, math.Max(math.Abs(got), math.Abs(want)))
}

func transpose(matrix qr.Matrix) qr.Matrix {
	result := make(qr.Matrix, len(matrix[0]))
	for column := range result {
		result[column] = make([]float64, len(matrix))
		for row := range matrix {
			result[column][row] = matrix[row][column]
		}
	}
	return result
}

func multiply(left, right qr.Matrix) qr.Matrix {
	result := make(qr.Matrix, len(left))
	for row := range result {
		result[row] = make([]float64, len(right[0]))
		for column := range result[row] {
			for index := range right {
				result[row][column] += left[row][index] * right[index][column]
			}
		}
	}
	return result
}

func identity(size int) qr.Matrix {
	result := make(qr.Matrix, size)
	for index := range result {
		result[index] = make([]float64, size)
		result[index][index] = 1
	}
	return result
}
