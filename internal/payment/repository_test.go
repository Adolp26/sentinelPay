package payment

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/adolp26/sentinelpay/internal/platform/database"
	"github.com/google/uuid"
)

func newTestRepository(t *testing.T) *Repository {
	t.Helper()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set; skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	return NewRepository(pool)
}

func TestRepository_CreateAndGetByID(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	p := validPayment()
	p.ID = uuid.NewString()

	if err := repo.Create(ctx, &p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.CreatedAt.IsZero() {
		t.Fatal("Create did not fill CreatedAt")
	}

	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.ID != p.ID || got.MerchantID != p.MerchantID || got.CustomerID != p.CustomerID ||
		got.Amount != p.Amount || got.Currency != p.Currency || got.Status != p.Status {
		t.Errorf("got %+v, want %+v", got, p)
	}
	if !got.CreatedAt.Equal(p.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, p.CreatedAt)
	}
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	repo := newTestRepository(t)

	for _, id := range []string{uuid.NewString(), "not-a-uuid"} {
		_, err := repo.GetByID(context.Background(), id)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("GetByID(%q) error = %v, want ErrNotFound", id, err)
		}
	}
}
