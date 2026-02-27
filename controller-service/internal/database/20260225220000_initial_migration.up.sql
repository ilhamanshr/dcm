CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS global_config (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    config JSONB NOT NULL,
    version BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

INSERT INTO global_config (config, version)
VALUES ('{}'::jsonb, 1)
ON CONFLICT (id) DO NOTHING;