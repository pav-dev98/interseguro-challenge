package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestDependenciesHealth(t *testing.T) {
	nodeAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer nodeAPI.Close()

	t.Setenv("NODE_API_URL", nodeAPI.URL)
	app := fiber.New()
	app.Get("/health/dependencies", dependenciesHealthHandler(nodeAPI.Client()))

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/health/dependencies", nil))
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	var payload struct {
		Status       string            `json:"status"`
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if payload.Status != "active" || payload.Dependencies["node-api"] != "active" {
		t.Fatalf("unexpected response: %#v", payload)
	}
}

func TestDependenciesHealthWhenNodeAPIIsUnavailable(t *testing.T) {
	t.Setenv("NODE_API_URL", "http://127.0.0.1:1")
	app := fiber.New()
	app.Get("/health/dependencies", dependenciesHealthHandler(&http.Client{Timeout: 50 * time.Millisecond}))

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/health/dependencies", nil))
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", response.StatusCode)
	}
}

func TestDependenciesHealthWhenNodeAPIReturnsAnError(t *testing.T) {
	nodeAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer nodeAPI.Close()

	t.Setenv("NODE_API_URL", nodeAPI.URL)
	app := fiber.New()
	app.Get("/health/dependencies", dependenciesHealthHandler(nodeAPI.Client()))

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/health/dependencies", nil))
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", response.StatusCode)
	}
}

func TestDependenciesHealthTimesOut(t *testing.T) {
	nodeAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer nodeAPI.Close()

	t.Setenv("NODE_API_URL", nodeAPI.URL)
	app := fiber.New()
	app.Get("/health/dependencies", dependenciesHealthHandler(&http.Client{Timeout: 10 * time.Millisecond}))

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/health/dependencies", nil))
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", response.StatusCode)
	}
}

func TestQREndpoint(t *testing.T) {
	app := fiber.New()
	app.Post("/qr", qrHandler)

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "valid matrix", body: `{"matrix":[[1,1],[1,0]]}`, wantStatus: http.StatusOK},
		{name: "invalid JSON", body: `{"matrix":`, wantStatus: http.StatusBadRequest},
		{name: "invalid matrix", body: `{"matrix":[[1,2],[3]]}`, wantStatus: http.StatusBadRequest},
		{name: "dependent columns", body: `{"matrix":[[1,2],[2,4]]}`, wantStatus: http.StatusUnprocessableEntity},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/qr", bytes.NewBufferString(test.body))
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("unexpected request error: %v", err)
			}
			defer response.Body.Close()

			if response.StatusCode != test.wantStatus {
				t.Fatalf("expected status %d, got %d", test.wantStatus, response.StatusCode)
			}
		})
	}
}
