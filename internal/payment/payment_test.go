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
		name      string
		modify    func(p *Payment)
		wantField string
		wantErr   error
	}{
		{"valid", func(p *Payment) {}, "", nil},
		{"missing merchant", func(p *Payment) { p.MerchantID = "" }, "merchant_id", ErrRequired},
		{"missing customer", func(p *Payment) { p.CustomerID = "" }, "customer_id", ErrRequired},
		{"zero amount", func(p *Payment) { p.Amount = 0 }, "amount", ErrInvalidAmount},
		{"negative amount", func(p *Payment) { p.Amount = -100 }, "amount", ErrInvalidAmount},
		{"empty currency", func(p *Payment) { p.Currency = "" }, "currency", ErrRequired},
		{"lowercase currency", func(p *Payment) { p.Currency = "brl" }, "currency", ErrInvalidCurrency},
		{"unknown currency", func(p *Payment) { p.Currency = "XYZ" }, "currency", ErrInvalidCurrency},
		{"malformed currency", func(p *Payment) { p.Currency = "R$" }, "currency", ErrInvalidCurrency},
		{"real but unsupported currency", func(p *Payment) { p.Currency = "JPY" }, "currency", ErrUnsupportedCurrency},
		{"iso special code", func(p *Payment) { p.Currency = "XXX" }, "currency", ErrUnsupportedCurrency},
		{"unknown status", func(p *Payment) { p.Status = "banana" }, "status", ErrInvalidStatus},
		{"empty status", func(p *Payment) { p.Status = "" }, "status", ErrInvalidStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validPayment()
			tt.modify(&p)

			err := p.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				return
			}

			var fe FieldError
			if !errors.As(err, &fe) {
				t.Fatalf("expected a FieldError, got %T", err)
			}
			if fe.Field != tt.wantField {
				t.Errorf("field = %q, want %q", fe.Field, tt.wantField)
			}
		})
	}
}

func TestPayment_Validate_ReturnsAllErrors(t *testing.T) {
	p := Payment{Currency: "JPY", Status: "banana"}

	err := p.Validate()

	for _, want := range []error{ErrRequired, ErrInvalidAmount, ErrUnsupportedCurrency, ErrInvalidStatus} {
		if !errors.Is(err, want) {
			t.Errorf("expected error %v in %v", want, err)
		}
	}

	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		t.Fatalf("expected joined errors, got %T", err)
	}
	if got := len(joined.Unwrap()); got != 5 {
		t.Errorf("got %d errors, want 5", got)
	}
}
