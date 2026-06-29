CREATE TABLE logistics_.assignments (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id       UUID NOT NULL REFERENCES logistics_.orders(id),
    driver_id      UUID NOT NULL,
    attempt_number INT NOT NULL DEFAULT 1,
    status         TEXT NOT NULL CHECK (status IN ('offered','accepted','rejected','timed_out')),
    offered_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    responded_at   TIMESTAMPTZ,
    timeout_at     TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_assignments_order ON logistics_.assignments(order_id);
CREATE INDEX idx_assignments_driver ON logistics_.assignments(driver_id);
