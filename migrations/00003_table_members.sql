-- +goose Up
CREATE TYPE RoleStatus AS ENUM ('OWNER', 'MEMBER');

CREATE TABLE IF NOT EXISTS members (
    id UUID PRIMARY KEY DEFAULT pg_catalog.gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,

    role RoleStatus NOT NULL DEFAULT 'MEMBER',

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose Down
DROP TABLE IF EXISTS members;
