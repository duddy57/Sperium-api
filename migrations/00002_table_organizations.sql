-- +goose Up
CREATE TYPE PlanStatus AS ENUM ('PRO', 'FREE');

CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(50) NOT NULL,
    domain VARCHAR(50) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,

    plan PlanStatus NOT NULL DEFAULT 'FREE',

    owner_id UUID REFERENCES users(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


-- +goose Down
DROP TABLE IF EXISTS organizations;