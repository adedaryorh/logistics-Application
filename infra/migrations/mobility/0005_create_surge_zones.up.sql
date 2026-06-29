CREATE TABLE mobility_.surge_zones (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    h3_cell       TEXT NOT NULL,
    multiplier    NUMERIC(4,2) NOT NULL,
    valid_from    TIMESTAMPTZ NOT NULL,
    valid_to      TIMESTAMPTZ NOT NULL
);
