// Package statistics provides the HTTP client for the Node statistics API.
package statistics

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/andrucar25/interseguro-coding-challenge/go-api/qr"
)

const statisticsPath = "/api/v1/statistics"

var (
	// ErrDownstream identifies an unavailable or invalid statistics response.
	ErrDownstream = errors.New("statistics downstream error")
	// ErrTimeout identifies a timeout while obtaining statistics.
	ErrTimeout = errors.New("statistics downstream timeout")
)

// Result is the statistics response returned by the Node API.
type Result struct {
	Maximum           float64 `json:"maximum"`
	Minimum           float64 `json:"minimum"`
	Sum               float64 `json:"sum"`
	Average           float64 `json:"average"`
	HasDiagonalMatrix bool    `json:"hasDiagonalMatrix"`
}

// Client calls the configured Node statistics endpoint using a shared HTTP client.
type Client struct {
	endpoint   string
	httpClient *http.Client
}

type request struct {
	Q qr.Matrix `json:"q"`
	R qr.Matrix `json:"r"`
}

type response struct {
	Maximum           *float64 `json:"maximum"`
	Minimum           *float64 `json:"minimum"`
	Sum               *float64 `json:"sum"`
	Average           *float64 `json:"average"`
	HasDiagonalMatrix *bool    `json:"hasDiagonalMatrix"`
}

// New constructs a client for an absolute HTTP(S) base URL.
func New(baseURL string, httpClient *http.Client) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, errors.New("node API URL must not be empty")
	}
	if httpClient == nil {
		return nil, errors.New("HTTP client must not be nil")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse node API URL: %w", err)
	}
	if (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		return nil, errors.New("node API URL must be an absolute HTTP(S) URL")
	}
	if parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		return nil, errors.New("node API URL must not contain a query or fragment")
	}

	endpoint, err := url.JoinPath(parsedURL.String(), statisticsPath)
	if err != nil {
		return nil, fmt.Errorf("build statistics endpoint: %w", err)
	}

	return &Client{endpoint: endpoint, httpClient: httpClient}, nil
}

// Calculate sends QR result matrices to Node and validates its statistics result.
func (c *Client) Calculate(ctx context.Context, q, r qr.Matrix) (Result, error) {
	body, err := json.Marshal(request{Q: q, R: r})
	if err != nil {
		return Result{}, fmt.Errorf("%w: encode statistics request: %w", ErrDownstream, err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return Result{}, fmt.Errorf("create statistics request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		if isTimeout(err) {
			return Result{}, fmt.Errorf("%w: request statistics: %w", ErrTimeout, err)
		}
		return Result{}, fmt.Errorf("%w: request statistics: %w", ErrDownstream, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, 1024))
		if readErr != nil {
			return Result{}, fmt.Errorf("%w: read statistics error response: %w", ErrDownstream, readErr)
		}
		return Result{}, fmt.Errorf("%w: statistics service returned status %d: %q", ErrDownstream, response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	result, err := decodeResponse(response.Body)
	if err != nil {
		return Result{}, fmt.Errorf("%w: decode statistics response: %w", ErrDownstream, err)
	}
	return result, nil
}

func decodeResponse(body io.Reader) (Result, error) {
	decoder := json.NewDecoder(body)
	var decoded response
	if err := decoder.Decode(&decoded); err != nil {
		return Result{}, err
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Result{}, errors.New("statistics response contains trailing JSON")
		}
		return Result{}, fmt.Errorf("read trailing statistics response JSON: %w", err)
	}

	if decoded.Maximum == nil || decoded.Minimum == nil || decoded.Sum == nil || decoded.Average == nil || decoded.HasDiagonalMatrix == nil {
		return Result{}, errors.New("statistics response is missing required fields")
	}
	if !isFinite(*decoded.Maximum) || !isFinite(*decoded.Minimum) || !isFinite(*decoded.Sum) || !isFinite(*decoded.Average) {
		return Result{}, errors.New("statistics response contains non-finite values")
	}

	return Result{
		Maximum:           *decoded.Maximum,
		Minimum:           *decoded.Minimum,
		Sum:               *decoded.Sum,
		Average:           *decoded.Average,
		HasDiagonalMatrix: *decoded.HasDiagonalMatrix,
	}, nil
}

func isTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var networkError net.Error
	return errors.As(err, &networkError) && networkError.Timeout()
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
