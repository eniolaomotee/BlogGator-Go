-- +goose Up
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    subscription_id UUID REFERENCES users_subscriptions(id),
    amount_cents INTEGER NOT NULL,
    currency TEXT NOT NULL DEFAULT 'usd',
    status TEXT NOT NULL , -- 'succeeded', 'failed', 'pending', 'refunded'
    stripe_payment_intent_id TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);


-- +goose Down
DROP TABLE payments;