CREATE TABLE logistics_.orders (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id          UUID NOT NULL,
    driver_id            UUID,
    merchant_id          UUID,
    type                 TEXT NOT NULL CHECK (type IN ('ride','food','parcel')),
    status               TEXT NOT NULL DEFAULT 'pending'
                         CHECK (status IN ('pending','awaiting_payment','paid','dispatching','assigned','picked_up','delivered','cancelled','failed')),
    pickup_lat           DOUBLE PRECISION NOT NULL,
    pickup_lng           DOUBLE PRECISION NOT NULL,
    pickup_h3            TEXT NOT NULL,
    dropoff_lat          DOUBLE PRECISION NOT NULL,
    dropoff_lng          DOUBLE PRECISION NOT NULL,
    dropoff_h3           TEXT NOT NULL,
    pickup_address       TEXT,
    dropoff_address      TEXT,
    price_minor          BIGINT,
    currency             TEXT NOT NULL DEFAULT 'NGN',
    surge_multiplier     NUMERIC(4,2) DEFAULT 1.00,
    idempotency_key      UUID NOT NULL UNIQUE,
    temporal_workflow_id TEXT,
    cancellation_reason  TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_orders_customer ON logistics_.orders(customer_id);
CREATE INDEX idx_orders_driver ON logistics_.orders(driver_id);
CREATE INDEX idx_orders_status ON logistics_.orders(status);
