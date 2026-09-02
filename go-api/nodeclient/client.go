package nodeclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

const RequestTimeout = 30 * time.Second

var (
	ErrNotConfigured    = errors.New("NODE_API_URL is not configured")
	ErrUnavailable      = errors.New("node-api is unavailable")
	ErrUnexpectedStatus = errors.New("node-api returned a non-success status")
	ErrInvalidResponse  = errors.New("node-api returned an invalid response")
)

type Statistics struct {
	Max               float64 `json:"max"`
	Min               float64 `json:"min"`
	Sum               float64 `json:"sum"`
	Average           float64 `json:"average"`
	HasDiagonalMatrix bool    `json:"hasDiagonalMatrix"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string, httpClient *http.Client) Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: RequestTimeout}
	}

	return Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (c Client) Statistics(ctx context.Context, q, r [][]float64) (Statistics, error) {
	if c.baseURL == "" {
		return Statistics{}, ErrNotConfigured
	}

	body, err := json.Marshal(struct {
		Q [][]float64 `json:"q"`
		R [][]float64 `json:"r"`
	}{Q: q, R: r})
	if err != nil {
		return Statistics{}, fmt.Errorf("%w: could not encode request", ErrUnavailable)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/statistics", bytes.NewReader(body))
	if err != nil {
		return Statistics{}, fmt.Errorf("%w: could not create request", ErrUnavailable)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return Statistics{}, fmt.Errorf("%w: request failed", ErrUnavailable)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return Statistics{}, fmt.Errorf("%w: %d", ErrUnexpectedStatus, response.StatusCode)
	}

	var payload struct {
		Max               *float64 `json:"max"`
		Min               *float64 `json:"min"`
		Sum               *float64 `json:"sum"`
		Average           *float64 `json:"average"`
		HasDiagonalMatrix *bool    `json:"hasDiagonalMatrix"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return Statistics{}, fmt.Errorf("%w: could not decode response", ErrInvalidResponse)
	}
	if payload.Max == nil || payload.Min == nil || payload.Sum == nil || payload.Average == nil || payload.HasDiagonalMatrix == nil ||
		!isFinite(*payload.Max) || !isFinite(*payload.Min) || !isFinite(*payload.Sum) || !isFinite(*payload.Average) {
		return Statistics{}, ErrInvalidResponse
	}

	return Statistics{
		Max:               *payload.Max,
		Min:               *payload.Min,
		Sum:               *payload.Sum,
		Average:           *payload.Average,
		HasDiagonalMatrix: *payload.HasDiagonalMatrix,
	}, nil
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
