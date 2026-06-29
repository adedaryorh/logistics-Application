CREATE TABLE operations_.feature_flags (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT NOT NULL UNIQUE,
    enabled      BOOLEAN NOT NULL DEFAULT FALSE,
    rollout_pct  INT NOT NULL DEFAULT 0 CHECK (rollout_pct BETWEEN 0 AND 100),
    description  TEXT,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
