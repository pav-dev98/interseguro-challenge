package nodeclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatisticsSendsMatricesAndDecodesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/statistics" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected JSON content type")
		}
		var payload struct {
			Q [][]float64 `json:"q"`
			R [][]float64 `json:"r"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("could not decode request: %v", err)
		}
		if payload.Q[0][0] != 1 || payload.R[0][0] != 2 {
			t.Fatalf("unexpected request payload: %#v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"max":2,"min":-1,"sum":3,"average":0.75,"hasDiagonalMatrix":true}`))
	}))
	defer server.Close()

	statistics, err := New(server.URL, server.Client()).Statistics(context.Background(), [][]float64{{1}}, [][]float64{{2}})
	if err != nil {
		t.Fatalf("Statistics returned an error: %v", err)
	}
	if statistics.Max != 2 || statistics.Min != -1 || statistics.Sum != 3 || statistics.Average != 0.75 || !statistics.HasDiagonalMatrix {
		t.Fatalf("unexpected statistics: %#v", statistics)
	}
}

func TestStatisticsReturnsExpectedErrors(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    error
	}{
		{name: "non-success status", handler: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }, want: ErrUnexpectedStatus},
		{name: "invalid JSON", handler: func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`not-json`)) }, want: ErrInvalidResponse},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			_, err := New(server.URL, server.Client()).Statistics(context.Background(), [][]float64{{1}}, [][]float64{{1}})
			if !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}

func TestStatisticsRequiresNodeAPIURL(t *testing.T) {
	_, err := New("", nil).Statistics(context.Background(), [][]float64{{1}}, [][]float64{{1}})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}
