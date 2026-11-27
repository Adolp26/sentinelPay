-- +goose Up
CREATE TABLE payments (
    id          UUID PRIMARY KEY,
    merchant_id TEXT        NOT NULL,
    customer_id TEXT        NOT NULL,
    amount      BIGINT      NOT NULL CHECK (amount > 0),
    currency    CHAR(3)     NOT NULL,
    status      TEXT        NOT NULL CHECK (status IN ('pending', 'approved', 'declined', 'review')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_merchant_created_at ON payments (merchant_id, created_at);

-- +goose Down
DROP TABLE payments;
