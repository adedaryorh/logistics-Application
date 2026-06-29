CREATE TABLE identity_.oauth_accounts (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES identity_.users(id) ON DELETE CASCADE,
    provider         TEXT NOT NULL CHECK (provider IN ('google','apple')),
    provider_user_id TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider, provider_user_id)
);
