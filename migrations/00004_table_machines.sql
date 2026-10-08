-- +goose Up
CREATE TABLE IF NOT EXISTS machines (
    id UUID PRIMARY KEY DEFAULT pg_catalog.gen_random_uuid(),

    organization_id UUID NOT NULL
        REFERENCES organizations(id) ON DELETE CASCADE,

    name TEXT NOT NULL,
    hostname TEXT,

    agent_version TEXT,
    last_seen_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT machines_organization_name_key UNIQUE (organization_id, name)
);
CREATE TABLE IF NOT EXISTS machine_enrollment_tokens (
    id            UUID PRIMARY KEY     DEFAULT pg_catalog.gen_random_uuid(),

    machine_id    UUID        NOT NULL
        REFERENCES machines (id) ON DELETE CASCADE,

    token_hash    TEXT        NOT NULL,
    used_at       TIMESTAMPTZ,
    expires_at    TIMESTAMPTZ NOT NULL,

    created_by_user_id UUID NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT machine_enrollment_tokens_token_hash_key UNIQUE (token_hash)
);

CREATE INDEX machines_organization_id_idx ON machines (organization_id);
CREATE INDEX machines_last_seen_at_idx ON machines (last_seen_at);
CREATE INDEX machine_enrollment_tokens_pending_expires_at_idx
    ON machine_enrollment_tokens (expires_at)
    WHERE used_at IS NULL;
-- +goose Down
DROP TABLE IF EXISTS machine_enrollment_tokens;
DROP TABLE IF EXISTS machines;
