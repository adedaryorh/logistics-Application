CREATE TABLE payment_.wallet_balances (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL UNIQUE,
    balance_minor BIGINT NOT NULL DEFAULT 0,
    currency      TEXT NOT NULL DEFAULT 'NGN',
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payment_.wallet_ledger (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id     UUID NOT NULL REFERENCES payment_.wallet_balances(id),
    type          TEXT NOT NULL CHECK (type IN ('credit','debit')),
    amount_minor  BIGINT NOT NULL,
    ref_id        UUID,
    description   TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
