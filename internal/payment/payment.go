package payment

import (
	"errors"
	"time"
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

var (
	ErrInvalidMerchantID = errors.New("merchant_id is required")
	ErrInvalidCustomerID = errors.New("customer_id is required")
	ErrInvalidAmount     = errors.New("amount must be positive")
	ErrInvalidCurrency   = errors.New("currency must be a 3-letter uppercase ISO code")
	ErrInvalidStatus     = errors.New("status is invalid")
)

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
	if p.MerchantID == "" {
		return ErrInvalidMerchantID
	}
	if p.CustomerID == "" {
		return ErrInvalidCustomerID
	}
	if p.Amount <= 0 {
		return ErrInvalidAmount
	}
	if !validCurrency(p.Currency) {
		return ErrInvalidCurrency
	}
	if !p.Status.Valid() {
		return ErrInvalidStatus
	}
	return nil
}

func validCurrency(c string) bool {
	if len(c) != 3 {
		return false
	}
	for _, r := range c {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
