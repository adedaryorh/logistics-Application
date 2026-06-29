CREATE TABLE logistics_.order_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES logistics_.orders(id),
    name        TEXT NOT NULL,
    quantity    INT NOT NULL DEFAULT 1,
    price_minor BIGINT NOT NULL,
    currency    TEXT NOT NULL
);
