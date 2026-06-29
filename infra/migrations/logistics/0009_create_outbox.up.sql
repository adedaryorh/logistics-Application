CREATE TABLE logistics_.outbox (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_id  UUID NOT NULL,
    event_type    TEXT NOT NULL,
    payload       JSONB NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at  TIMESTAMPTZ
);
CREATE INDEX idx_outbox_unpublished ON logistics_.outbox(created_at)
    WHERE published_at IS NULL;
