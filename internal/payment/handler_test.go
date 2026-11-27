package payment

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeStore struct {
	payments  map[string]Payment
	createErr error
}

func newFakeStore() *fakeStore {
	return &fakeStore{payments: map[string]Payment{}}
}

func (f *fakeStore) Create(_ context.Context, p *Payment) error {
	if f.createErr != nil {
		return f.createErr
	}
	p.CreatedAt = time.Now().UTC()
	f.payments[p.ID] = *p
	return nil
}

func (f *fakeStore) GetByID(_ context.Context, id string) (Payment, error) {
	p, ok := f.payments[id]
	if !ok {
		return Payment{}, ErrNotFound
	}
	return p, nil
}

func newTestServer(s store) *http.ServeMux {
	mux := http.NewServeMux()
	NewHandler(s, slog.New(slog.NewTextHandler(io.Discard, nil))).RegisterRoutes(mux)
	return mux
}

func do(mux http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

const validBody = `{"merchant_id":"m1","customer_id":"c1","amount":1000,"currency":"BRL"}`

func TestCreatePayment_Success(t *testing.T) {
	s := newFakeStore()
	rec := do(newTestServer(s), http.MethodPost, "/v1/payments", validBody)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body)
	}

	var got Payment
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID == "" || got.Status != StatusPending || got.CreatedAt.IsZero() {
		t.Errorf("unexpected payment: %+v", got)
	}
	if _, ok := s.payments[got.ID]; !ok {
		t.Error("payment was not stored")
	}
}

func TestCreatePayment_BadRequests(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantError  string
		wantFields []string
	}{
		{"malformed json", `{"merchant_id":`, "invalid_json", nil},
		{"unknown field", `{"merchant_id":"m1","status":"approved"}`, "invalid_json", nil},
		{"wrong type", `{"amount":"1000"}`, "invalid_json", nil},
		{"all fields invalid", `{"currency":"JPY"}`, "validation_failed", []string{"merchant_id", "customer_id", "amount", "currency"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(newTestServer(newFakeStore()), http.MethodPost, "/v1/payments", tt.body)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}

			var got errorResponse
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got.Error != tt.wantError {
				t.Errorf("error = %q, want %q", got.Error, tt.wantError)
			}
			if len(got.Fields) != len(tt.wantFields) {
				t.Fatalf("fields = %+v, want %v", got.Fields, tt.wantFields)
			}
			for i, f := range tt.wantFields {
				if got.Fields[i].Field != f {
					t.Errorf("fields[%d] = %q, want %q", i, got.Fields[i].Field, f)
				}
			}
		})
	}
}

func TestCreatePayment_StoreError(t *testing.T) {
	s := newFakeStore()
	s.createErr = errors.New("connection reset")

	rec := do(newTestServer(s), http.MethodPost, "/v1/payments", validBody)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "connection reset") {
		t.Error("internal error leaked to the client")
	}
}

func TestGetPayment(t *testing.T) {
	s := newFakeStore()
	s.payments["abc"] = Payment{ID: "abc", Status: StatusPending}
	mux := newTestServer(s)

	if rec := do(mux, http.MethodGet, "/v1/payments/abc", ""); rec.Code != http.StatusOK {
		t.Errorf("existing: status = %d, want 200", rec.Code)
	}
	if rec := do(mux, http.MethodGet, "/v1/payments/missing", ""); rec.Code != http.StatusNotFound {
		t.Errorf("missing: status = %d, want 404", rec.Code)
	}
}
