CREATE TABLE identity_.refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES identity_.users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    family_id  UUID NOT NULL,
    device_id  TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rt_user_id ON identity_.refresh_tokens(user_id);
CREATE INDEX idx_rt_family ON identity_.refresh_tokens(family_id);
