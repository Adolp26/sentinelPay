package payment

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/adolp26/sentinelpay/internal/platform/httpx"
	"github.com/google/uuid"
)

const maxBodyBytes = 1 << 20

type store interface {
	Create(ctx context.Context, p *Payment) error
	GetByID(ctx context.Context, id string) (Payment, error)
}

type Handler struct {
	store  store
	logger *slog.Logger
}

func NewHandler(s store, logger *slog.Logger) *Handler {
	return &Handler{store: s, logger: logger}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/payments", h.create)
	mux.HandleFunc("GET /v1/payments/{id}", h.get)
}

type createPaymentRequest struct {
	MerchantID string `json:"merchant_id"`
	CustomerID string `json:"customer_id"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
}

type fieldErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error  string               `json:"error"`
	Fields []fieldErrorResponse `json:"fields,omitempty"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req createPaymentRequest
	if err := dec.Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	p := Payment{
		ID:         uuid.NewString(),
		MerchantID: req.MerchantID,
		CustomerID: req.CustomerID,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Status:     StatusPending,
	}

	if err := p.Validate(); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, errorResponse{
			Error:  "validation_failed",
			Fields: fieldErrors(err),
		})
		return
	}

	if err := h.store.Create(r.Context(), &p); err != nil {
		h.logger.Error("create payment failed", "error", err)
		httpx.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal_error"})
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, p)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	p, err := h.store.GetByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, ErrNotFound) {
		httpx.WriteJSON(w, http.StatusNotFound, errorResponse{Error: "not_found"})
		return
	}
	if err != nil {
		h.logger.Error("get payment failed", "error", err)
		httpx.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal_error"})
		return
	}

	httpx.WriteJSON(w, http.StatusOK, p)
}

func fieldErrors(err error) []fieldErrorResponse {
	var errs []error
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		errs = joined.Unwrap()
	} else {
		errs = []error{err}
	}

	out := make([]fieldErrorResponse, 0, len(errs))
	for _, e := range errs {
		var fe FieldError
		if errors.As(e, &fe) {
			out = append(out, fieldErrorResponse{Field: fe.Field, Message: fe.Err.Error()})
		}
	}
	return out
}
