CREATE TABLE payment_.audit_log (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID REFERENCES payment_.transactions(id),
    action         TEXT NOT NULL,
    actor_id       UUID,
    actor_type     TEXT,
    metadata       JSONB,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
