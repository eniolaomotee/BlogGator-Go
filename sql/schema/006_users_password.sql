-- +goose Up
ALTER TABLE users ADD COLUMN password_hash TEXT NOT NULL DEFAULT '';


CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash TEXT NOT NULL,
    name TEXT NOT NULL,
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP
);