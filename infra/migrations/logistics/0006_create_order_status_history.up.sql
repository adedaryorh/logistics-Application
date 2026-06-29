CREATE TABLE logistics_.order_status_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES logistics_.orders(id),
    from_status TEXT,
    to_status   TEXT NOT NULL,
    reason      TEXT,
    actor_id    UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
