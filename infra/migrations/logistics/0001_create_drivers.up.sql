CREATE TABLE logistics_.drivers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL UNIQUE,
    full_name   TEXT NOT NULL,
    phone       TEXT NOT NULL,
    type        TEXT NOT NULL CHECK (type IN ('ride','food','parcel')),
    lat         DOUBLE PRECISION NOT NULL,
    lng         DOUBLE PRECISION NOT NULL,
    h3_cell     TEXT NOT NULL,
    rating      NUMERIC(3,2) NOT NULL DEFAULT 5.00,
    total_trips INT NOT NULL DEFAULT 0,
    online      BOOLEAN NOT NULL DEFAULT FALSE,
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
