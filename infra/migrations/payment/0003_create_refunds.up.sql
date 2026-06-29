CREATE TABLE payment_.refunds (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id      UUID NOT NULL REFERENCES payment_.transactions(id),
    amount_minor        BIGINT NOT NULL,
    reason              TEXT,
    status              TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending','completed','failed')),
    provider_refund_id  TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
