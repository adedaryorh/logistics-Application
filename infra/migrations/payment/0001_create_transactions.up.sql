CREATE TABLE payment_.transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL,
    customer_id     UUID NOT NULL,
    amount_minor    BIGINT NOT NULL,
    currency        TEXT NOT NULL DEFAULT 'NGN',
    provider        TEXT NOT NULL CHECK (provider IN ('flutterwave','paystack','stripe')),
    provider_tx_id  TEXT,
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','initialized','completed','failed','refunded')),
    checkout_url    TEXT,
    idempotency_key UUID NOT NULL UNIQUE,
    metadata        JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
