CREATE TABLE identity_.oauth_sessions (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider              TEXT NOT NULL CHECK (provider IN ('google','apple')),
    state                 TEXT NOT NULL UNIQUE,
    nonce                 TEXT NOT NULL,
    redirect_uri          TEXT NOT NULL,
    code_challenge        TEXT,
    code_challenge_method TEXT CHECK (code_challenge_method IN ('S256', 'plain')),
    expires_at            TIMESTAMPTZ NOT NULL,
    used_at               TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_oauth_sessions_provider_state ON identity_.oauth_sessions(provider, state);
