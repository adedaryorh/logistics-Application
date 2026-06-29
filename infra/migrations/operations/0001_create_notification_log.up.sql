CREATE TABLE operations_.notification_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    type            TEXT NOT NULL CHECK (type IN ('email','push','sms')),
    template_name   TEXT NOT NULL,
    recipient       TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','sent','failed','bounced')),
    provider        TEXT,
    provider_msg_id TEXT,
    error_message   TEXT,
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
