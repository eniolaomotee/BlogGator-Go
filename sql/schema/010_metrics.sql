-- +goose Up
CREATE TABLE usage_metrics(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    metric_type TEXT NOT NULL, -- 'feeds', 'posts', 'api_calls'
    count INTEGER NOT NULL,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);


-- +goose Down
DROP TABLE usage_metrics;