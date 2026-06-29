CREATE TABLE identity_.users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email          TEXT UNIQUE NOT NULL,
    phone          TEXT UNIQUE,
    password_hash  TEXT,
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    role           TEXT NOT NULL DEFAULT 'customer'
                   CHECK (role IN ('customer','driver','merchant','admin','ops','mcp_agent')),
    status         TEXT NOT NULL DEFAULT 'active'
                   CHECK (status IN ('active','suspended','deleted')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON identity_.users(email) WHERE deleted_at IS NULL;
