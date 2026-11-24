package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func slowHandler(t *testing.T, gotErr *error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(1 * time.Second):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			*gotErr = r.Context().Err()
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}
}

func TestWithTimeout_CancelsSlowRequest(t *testing.T) {
	var gotErr error
	h := withTimeout(50*time.Millisecond, slowHandler(t, &gotErr))

	rec := httptest.NewRecorder()
	start := time.Now()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if !errors.Is(gotErr, context.DeadlineExceeded) {
		t.Fatalf("esperava DeadlineExceeded, veio %v", gotErr)
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("esperava 503, veio %d", rec.Code)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("deveria ter parado em ~50ms, levou %v", elapsed)
	}
}

func TestWithTimeout_FastRequestPasses(t *testing.T) {
	h := withTimeout(time.Second, http.HandlerFunc(healthHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, veio %d", rec.Code)
	}
}
