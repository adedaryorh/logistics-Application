CREATE TABLE mobility_.location_events (
    driver_id    UUID NOT NULL,
    lat          DOUBLE PRECISION NOT NULL,
    lng          DOUBLE PRECISION NOT NULL,
    h3_cell_r8   TEXT NOT NULL,
    h3_cell_r9   TEXT NOT NULL,
    accuracy_m   FLOAT,
    speed_kmh    FLOAT,
    heading_deg  FLOAT,
    recorded_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
