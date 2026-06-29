CREATE TABLE mobility_.route_cache (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_h3_r9   TEXT NOT NULL,
    to_h3_r9     TEXT NOT NULL,
    route_json   JSONB NOT NULL,
    cached_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(from_h3_r9, to_h3_r9)
);
