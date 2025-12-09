-- +goose Up
CREATE TABLE subscription_tiers(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    price_cents INTEGER NOT NULL,
    max_feeds INTEGER NOT NULL,
    max_posts INTEGER,
    features JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE subscription_tiers;