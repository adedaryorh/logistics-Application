CREATE TABLE logistics_.ratings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES logistics_.orders(id),
    rater_id    UUID NOT NULL,
    ratee_id    UUID NOT NULL,
    ratee_type  TEXT NOT NULL CHECK (ratee_type IN ('driver','customer')),
    score       INT NOT NULL CHECK (score BETWEEN 1 AND 5),
    comment     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(order_id, rater_id, ratee_id)
);
