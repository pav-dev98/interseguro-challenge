package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/interseguro/challenge/go-api/nodeclient"
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
	statisticsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"max":1,"min":-1,"sum":0,"average":0,"hasDiagonalMatrix":false}`))
	}))
	defer statisticsServer.Close()

	app := fiber.New()
	app.Post("/qr", qrHandler(nodeclient.New(statisticsServer.URL, statisticsServer.Client())))

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

func TestQREndpointStatisticsIntegration(t *testing.T) {
	tests := []struct {
		name       string
		server     http.HandlerFunc
		baseURL    string
		client     *http.Client
		wantStatus int
	}{
		{
			name: "returns QR and statistics",
			server: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/statistics" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"max":1.4142135623730951,"min":-0.7071067811865475,"sum":3.414213562373095,"average":0.42677669529663687,"hasDiagonalMatrix":false}`))
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "node unavailable",
			baseURL:    "http://127.0.0.1:1",
			client:     &http.Client{Timeout: 50 * time.Millisecond},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "node URL is not configured",
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "node returns error",
			server: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantStatus: http.StatusBadGateway,
		},
		{
			name: "node returns invalid JSON",
			server: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("not-json"))
			},
			wantStatus: http.StatusBadGateway,
		},
		{
			name: "node times out",
			server: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			},
			client:     &http.Client{Timeout: 10 * time.Millisecond},
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			baseURL := test.baseURL
			client := test.client
			var server *httptest.Server
			if test.server != nil {
				server = httptest.NewServer(test.server)
				defer server.Close()
				baseURL = server.URL
				if client == nil {
					client = server.Client()
				}
			}

			app := fiber.New()
			app.Post("/qr", qrHandler(nodeclient.New(baseURL, client)))
			request := httptest.NewRequest(http.MethodPost, "/qr", bytes.NewBufferString(`{"matrix":[[1,1],[1,0]]}`))
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("unexpected request error: %v", err)
			}
			defer response.Body.Close()

			if response.StatusCode != test.wantStatus {
				t.Fatalf("expected status %d, got %d", test.wantStatus, response.StatusCode)
			}

			if test.wantStatus == http.StatusOK {
				var payload struct {
					Q          [][]float64           `json:"q"`
					R          [][]float64           `json:"r"`
					Statistics nodeclient.Statistics `json:"statistics"`
				}
				if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
					t.Fatalf("could not decode response: %v", err)
				}
				if len(payload.Q) != 2 || len(payload.R) != 2 || payload.Statistics.Max != 1.4142135623730951 {
					t.Fatalf("unexpected payload: %#v", payload)
				}
			}
		})
	}
}
