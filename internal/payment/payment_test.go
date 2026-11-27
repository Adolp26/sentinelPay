package payment

import (
	"errors"
	"testing"
)

func validPayment() Payment {
	return Payment{
		MerchantID: "merchant-1",
		CustomerID: "customer-1",
		Amount:     1000,
		Currency:   "BRL",
		Status:     StatusPending,
	}
}

func TestPayment_Validate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(p *Payment)
		wantErr error
	}{
		{"valid", func(p *Payment) {}, nil},
		{"missing merchant", func(p *Payment) { p.MerchantID = "" }, ErrInvalidMerchantID},
		{"missing customer", func(p *Payment) { p.CustomerID = "" }, ErrInvalidCustomerID},
		{"zero amount", func(p *Payment) { p.Amount = 0 }, ErrInvalidAmount},
		{"negative amount", func(p *Payment) { p.Amount = -100 }, ErrInvalidAmount},
		{"empty currency", func(p *Payment) { p.Currency = "" }, ErrInvalidCurrency},
		{"lowercase currency", func(p *Payment) { p.Currency = "brl" }, ErrInvalidCurrency},
		{"currency too long", func(p *Payment) { p.Currency = "BRLX" }, ErrInvalidCurrency},
		{"unknown status", func(p *Payment) { p.Status = "banana" }, ErrInvalidStatus},
		{"empty status", func(p *Payment) { p.Status = "" }, ErrInvalidStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validPayment()
			tt.modify(&p)

			err := p.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
