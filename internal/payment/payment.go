package payment

import (
	"errors"
	"time"

	"golang.org/x/text/currency"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusDeclined Status = "declined"
	StatusReview   Status = "review"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusApproved, StatusDeclined, StatusReview:
		return true
	}
	return false
}

var supportedCurrencies = map[currency.Unit]bool{
	currency.BRL: true,
	currency.USD: true,
	currency.EUR: true,
}

var (
	ErrRequired            = errors.New("is required")
	ErrInvalidAmount       = errors.New("must be positive")
	ErrInvalidCurrency     = errors.New("must be an uppercase ISO 4217 code")
	ErrUnsupportedCurrency = errors.New("is not supported")
	ErrInvalidStatus       = errors.New("is invalid")
)

type FieldError struct {
	Field string
	Err   error
}

func (e FieldError) Error() string { return e.Field + " " + e.Err.Error() }

func (e FieldError) Unwrap() error { return e.Err }

type Payment struct {
	ID         string    `json:"id"`
	MerchantID string    `json:"merchant_id"`
	CustomerID string    `json:"customer_id"`
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	Status     Status    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

func (p Payment) Validate() error {
	var errs []error

	if p.MerchantID == "" {
		errs = append(errs, FieldError{"merchant_id", ErrRequired})
	}
	if p.CustomerID == "" {
		errs = append(errs, FieldError{"customer_id", ErrRequired})
	}
	if p.Amount <= 0 {
		errs = append(errs, FieldError{"amount", ErrInvalidAmount})
	}
	if err := validateCurrency(p.Currency); err != nil {
		errs = append(errs, FieldError{"currency", err})
	}
	if !p.Status.Valid() {
		errs = append(errs, FieldError{"status", ErrInvalidStatus})
	}

	return errors.Join(errs...)
}

func validateCurrency(code string) error {
	if code == "" {
		return ErrRequired
	}
	unit, err := currency.ParseISO(code)
	if err != nil || unit.String() != code {
		return ErrInvalidCurrency
	}
	if !supportedCurrencies[unit] {
		return ErrUnsupportedCurrency
	}
	return nil
}
