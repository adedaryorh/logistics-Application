ALTER TABLE logistics_.orders DROP CONSTRAINT orders_type_check;
ALTER TABLE logistics_.orders ADD CONSTRAINT orders_type_check CHECK (type IN ('ride','food','parcel','agricultural'));
ALTER TABLE logistics_.orders
    ADD COLUMN IF NOT EXISTS platform_user_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS source_platform_service VARCHAR(32),
    ADD COLUMN IF NOT EXISTS farmsense_request_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS marketplace_request_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS agricultural_shipment JSONB;

CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_marketplace_request_id
    ON logistics_.orders (marketplace_request_id) WHERE marketplace_request_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS logistics_.agricultural_quotes (
    id UUID PRIMARY KEY,
    platform_service VARCHAR(32) NOT NULL CHECK (platform_service IN ('farmsense', 'taskam')),
    platform_user_id VARCHAR(128) NOT NULL,
    farmsense_request_id VARCHAR(128) NOT NULL,
    marketplace_request_id VARCHAR(128) NOT NULL,
    idempotency_key VARCHAR(128) NOT NULL,
    pickup JSONB NOT NULL,
    dropoff JSONB NOT NULL,
    shipment JSONB NOT NULL,
    price_minor BIGINT NOT NULL CHECK (price_minor > 0),
    currency CHAR(3) NOT NULL DEFAULT 'NGN',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (platform_service, idempotency_key)
);

CREATE TABLE IF NOT EXISTS logistics_.agricultural_booking_idempotency (
    platform_service VARCHAR(32) NOT NULL,
    idempotency_key VARCHAR(128) NOT NULL,
    order_id UUID NOT NULL REFERENCES logistics_.orders(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (platform_service, idempotency_key)
);

CREATE TABLE IF NOT EXISTS logistics_.delivery_proofs (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES logistics_.orders(id),
    driver_id UUID NOT NULL REFERENCES logistics_.drivers(id),
    proof_type VARCHAR(16) NOT NULL CHECK (proof_type IN ('pickup', 'delivery')),
    evidence_url TEXT NOT NULL,
    notes TEXT,
    recipient_name VARCHAR(160),
    coordinate JSONB NOT NULL,
    captured_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (order_id, proof_type)
);

CREATE TABLE IF NOT EXISTS logistics_.outbound_status_webhooks (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES logistics_.orders(id),
    event_type VARCHAR(96) NOT NULL,
    payload JSONB NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMPTZ,
    dead_lettered_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_outbound_status_webhooks_pending ON logistics_.outbound_status_webhooks (next_attempt_at) WHERE delivered_at IS NULL;
