package statistics

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/andrucar25/interseguro-coding-challenge/go-api/qr"
)

func TestCalculateSendsNodeContractAndDecodesResult(t *testing.T) {
	q := qr.Matrix{{1}, {0}}
	r := qr.Matrix{{2}}
	want := Result{Maximum: 2, Minimum: 0, Sum: 3, Average: 1, HasDiagonalMatrix: true}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			t.Errorf("method = %s, want %s", req.Method, http.MethodPost)
		}
		if req.URL.Path != statisticsPath {
			t.Errorf("path = %q, want %q", req.URL.Path, statisticsPath)
		}
		if contentType := req.Header.Get("Content-Type"); contentType != "application/json" {
			t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
		}

		var got request
		if err := json.NewDecoder(req.Body).Decode(&got); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if !matricesEqual(got.Q, q) || !matricesEqual(got.R, r) {
			t.Errorf("request = %#v, want q=%#v r=%#v", got, q, r)
		}

		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(want); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client := newClient(t, server.URL, server.Client())
	got, err := client.Calculate(context.Background(), q, r)
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}
	if got != want {
		t.Errorf("Calculate() = %#v, want %#v", got, want)
	}
}

func TestCalculateRejectsDownstreamFailures(t *testing.T) {
	testCases := []struct {
		name string
		body string
		code int
	}{
		{name: "Node client error", code: http.StatusBadRequest, body: `{"error":"invalid_request"}`},
		{name: "Node server error", code: http.StatusInternalServerError, body: `failure`},
		{name: "malformed JSON", code: http.StatusOK, body: `{"maximum":`},
		{name: "trailing JSON", code: http.StatusOK, body: validResponse + ` {}`},
		{name: "missing required field", code: http.StatusOK, body: `{"maximum":2,"minimum":0,"sum":2,"average":1}`},
		{name: "wrong required field type", code: http.StatusOK, body: `{"maximum":2,"minimum":0,"sum":2,"average":1,"hasDiagonalMatrix":"yes"}`},
		{name: "non-finite aggregate", code: http.StatusOK, body: `{"maximum":1e9999,"minimum":0,"sum":2,"average":1,"hasDiagonalMatrix":false}`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(testCase.code)
				_, _ = io.WriteString(writer, testCase.body)
			}))
			defer server.Close()

			client := newClient(t, server.URL, server.Client())
			_, err := client.Calculate(context.Background(), qr.Matrix{{1}}, qr.Matrix{{1}})
			if !errors.Is(err, ErrDownstream) {
				t.Fatalf("Calculate() error = %v, want errors.Is(..., ErrDownstream)", err)
			}
		})
	}
}

func TestCalculateCategorizesDeadlineAsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()

	client := newClient(t, server.URL, server.Client())
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := client.Calculate(ctx, qr.Matrix{{1}}, qr.Matrix{{1}})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("Calculate() error = %v, want errors.Is(..., ErrTimeout)", err)
	}
}

func TestCalculateCategorizesNetworkFailureAsDownstreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	client := newClient(t, server.URL, server.Client())
	server.Close()

	_, err := client.Calculate(context.Background(), qr.Matrix{{1}}, qr.Matrix{{1}})
	if !errors.Is(err, ErrDownstream) {
		t.Fatalf("Calculate() error = %v, want errors.Is(..., ErrDownstream)", err)
	}
}

func TestCalculateClosesResponseBody(t *testing.T) {
	body := &trackingBody{Reader: strings.NewReader(validResponse)}
	client := newClient(t, "http://example.test", &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
			Header:     make(http.Header),
		}, nil
	})})

	if _, err := client.Calculate(context.Background(), qr.Matrix{{1}}, qr.Matrix{{1}}); err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}
	if !body.closed {
		t.Error("response body was not closed")
	}
}

func TestNewRejectsInvalidConfiguration(t *testing.T) {
	testCases := []string{"", "relative/path", "ftp://example.test", "http://", "http://example.test?query=value"}
	for _, baseURL := range testCases {
		t.Run(baseURL, func(t *testing.T) {
			if _, err := New(baseURL, &http.Client{}); err == nil {
				t.Fatalf("New(%q) error = nil, want an error", baseURL)
			}
		})
	}
}

func newClient(t *testing.T, baseURL string, httpClient *http.Client) *Client {
	t.Helper()
	client, err := New(baseURL, httpClient)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return client
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

const validResponse = `{"maximum":2,"minimum":0,"sum":2,"average":1,"hasDiagonalMatrix":true}`

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

type trackingBody struct {
	io.Reader
	closed bool
}

func (body *trackingBody) Close() error {
	body.closed = true
	return nil
}
