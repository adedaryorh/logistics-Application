CREATE TABLE logistics_.menu_items (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id  UUID NOT NULL REFERENCES logistics_.merchants(id),
    name         TEXT NOT NULL,
    description  TEXT,
    price_minor  BIGINT NOT NULL,
    currency     TEXT NOT NULL DEFAULT 'NGN',
    is_available BOOLEAN DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
