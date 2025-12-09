-- name: UserTestTiers :one
INSERT INTO subscription_tiers(name, price_cents, max_feeds, max_posts, features) VALUES
('free', 0, 5, 100, '{"api_access": false, "tui_access": false}'),
('pro', 900, 50, NULL, '{"api_access": true, "tui_access": true, "email_support": true}'),
('team', 2900, 200, NULL, '{"api_access": true, "tui_access": true, "multi_user": true, "webhooks": true}'),
('enterprise', 9900, NULL, NULL, '{"api_access": true, "tui_access": true, "multi_user": true, "webhooks": true, "sla": true}')
RETURNING *;

-- ==========================================
-- SUBSCRIPTION TIERS
-- ==========================================

-- name: GetAllTiers :many
SELECT * FROM subscription_tiers
ORDER BY price_cents ASC;

-- name: GetTierByID :one
SELECT * FROM subscription_tiers
WHERE id = $1;

-- name: GetTierByName :one
SELECT * FROM subscription_tiers
WHERE name = $1;

-- name: CreateTier :one
INSERT INTO subscription_tiers (
    id,
    name,
    price_cents,
    max_feeds,
    max_posts,
    features,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateTier :one
UPDATE subscription_tiers
SET 
    price_cents = $2,
    max_feeds = $3,
    max_posts = $4,
    features = $5
WHERE id = $1
RETURNING *;

-- ==========================================
-- USER SUBSCRIPTIONS
-- ==========================================

-- name: GetUserSubscription :one
SELECT * FROM users_subscriptions
WHERE user_id = $1
AND status IN ('active', 'trialing')
ORDER BY created_at DESC
LIMIT 1;

-- name: GetSubscriptionByID :one
SELECT * FROM users_subscriptions
WHERE id = $1;

-- name: GetSubscriptionByStripeID :one
SELECT * FROM users_subscriptions
WHERE stripe_subscription_id = $1;

-- name: CreateSubscription :one
INSERT INTO users_subscriptions (
    id,
    user_id,
    tier_id,
    status,
    current_period_start,
    current_period_end,
    stripe_customer_id,
    stripe_subscription_id,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdateSubscriptionStatus :exec
UPDATE users_subscriptions
SET 
    status = $2,
    updated_at = $3
WHERE id = $1;

-- name: UpdateSubscriptionPeriod :exec
UPDATE users_subscriptions
SET 
    current_period_start = $2,
    current_period_end = $3,
    updated_at = $4
WHERE id = $1;

-- name: UpdateSubscriptionCancelation :exec
UPDATE users_subscriptions
SET 
    cancel_at_end_period= $2,
    updated_at = $3
WHERE id = $1;

-- name: UpdateSubscriptionTier :exec
UPDATE users_subscriptions
SET 
    tier_id = $2,
    updated_at = $3
WHERE id = $1;

-- name: CancelSubscription :exec
UPDATE users_subscriptions
SET 
    status = 'canceled',
    updated_at = $2
WHERE id = $1;

-- name: GetUserSubscriptionHistory :many
SELECT * FROM users_subscriptions
WHERE user_id = $1
ORDER BY created_at DESC;

-- ==========================================
-- PAYMENTS
-- ==========================================

-- name: CreatePayment :one
INSERT INTO payments (
    id,
    user_id,
    subscription_id,
    amount_cents,
    currency,
    status,
    stripe_payment_intent_id,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetPaymentByStripeID :one
SELECT * FROM payments
WHERE stripe_payment_intent_id = $1;

-- name: UpdatePaymentStatus :exec
UPDATE payments
SET status = $2
WHERE id = $1;

-- name: GetUserPayments :many
SELECT * FROM payments
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: GetPaymentsBySubscription :many
SELECT * FROM payments
WHERE subscription_id = $1
ORDER BY created_at DESC;

-- name: GetSuccessfulPaymentsTotal :one
SELECT COALESCE(SUM(amount_cents), 0) as total
FROM payments
WHERE user_id = $1
AND status = 'succeeded';

-- ==========================================
-- USAGE METRICS
-- ==========================================

-- name: CreateUsageMetric :one
INSERT INTO usage_metrics (
    id,
    user_id,
    metric_type,
    count,
    period_start,
    period_end,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetCurrentPeriodUsage :one
SELECT * FROM usage_metrics
WHERE user_id = $1
AND metric_type = $2
AND period_end > NOW()
ORDER BY created_at DESC
LIMIT 1;

-- name: IncrementUsageMetric :exec
UPDATE usage_metrics
SET count = count + 1
WHERE id = $1;

-- name: GetUsageHistory :many
SELECT * FROM usage_metrics
WHERE user_id = $1
AND metric_type = $2
ORDER BY period_start DESC
LIMIT $3;

-- ==========================================
-- USAGE LIMITS & COUNTS
-- ==========================================

-- name: CountUserFeeds :one
SELECT COUNT(*) FROM feed_follows
WHERE user_id = $1;

-- name: CountUserPosts :one
SELECT COUNT(*) 
FROM posts p
JOIN feed_follows ff ON p.feed_id = ff.feed_id
WHERE ff.user_id = $1;

-- name: GetAPIUsageThisMonth :one
SELECT COALESCE(count, 0) as count
FROM usage_metrics
WHERE user_id = $1
AND metric_type = 'api_calls'
AND period_start >= date_trunc('month', NOW())
AND period_end < date_trunc('month', NOW() + interval '1 month')
ORDER BY created_at DESC
LIMIT 1;

-- name: IncrementAPIUsage :exec
INSERT INTO usage_metrics (
    id,
    user_id,
    metric_type,
    count,
    period_start,
    period_end,
    created_at
) VALUES (
    gen_random_uuid(),
    $1,
    'api_calls',
    1,
    date_trunc('month', NOW()),
    date_trunc('month', NOW() + interval '1 month'),
    NOW()
)
ON CONFLICT (user_id, metric_type, period_start)
DO UPDATE SET count = usage_metrics.count + 1;

-- ==========================================
-- SUBSCRIPTION CHECKS
-- ==========================================

-- name: CheckUserHasFeature :one
SELECT 
    CASE 
        WHEN us.id IS NULL THEN false
        WHEN st.features->$2 IS NULL THEN false
        WHEN (st.features->>$2)::boolean = true THEN true
        ELSE false
    END as has_feature
FROM users u
LEFT JOIN users_subscriptions us ON u.id = us.user_id 
    AND us.status IN ('active', 'trialing')
LEFT JOIN subscription_tiers st ON us.tier_id = st.id
WHERE u.id = $1
ORDER BY us.created_at DESC
LIMIT 1;

-- name: GetUserTierLimits :one
SELECT 
    COALESCE(st.max_feeds, 2147483647) as max_feeds,
    COALESCE(st.max_posts, 2147483647) as max_posts,
    st.features
FROM users u
LEFT JOIN users_subscriptions us ON u.id = us.user_id 
    AND us.status IN ('active', 'trialing')
LEFT JOIN subscription_tiers st ON us.tier_id = st.id
WHERE u.id = $1
ORDER BY us.created_at DESC
LIMIT 1;

-- name: CanUserAddFeed :one
SELECT 
    CASE 
        WHEN st.max_feeds IS NULL THEN true
                WHEN (SELECT COUNT(*) FROM feed_follows WHERE feed_follows.user_id = $1) < st.max_feeds THEN true
        ELSE false
    END as can_add
FROM users u
LEFT JOIN users_subscriptions us ON u.id = us.user_id 
    AND us.status IN ('active', 'trialing')
LEFT JOIN subscription_tiers st ON us.tier_id = st.id
WHERE u.id = $1
ORDER BY us.created_at DESC
LIMIT 1;

-- ==========================================
-- STRIPE CUSTOMER MANAGEMENT
-- ==========================================

-- name: UpdateUserStripeCustomer :exec
UPDATE users_subscriptions
SET 
    stripe_customer_id = $2,
    updated_at = $3
WHERE user_id = $1;

-- name: GetUserByStripeCustomerID :one
SELECT u.* 
FROM users u
JOIN users_subscriptions us ON u.id = us.user_id
WHERE us.stripe_customer_id = $1
LIMIT 1;

-- ==========================================
-- ANALYTICS & REPORTING
-- ==========================================

-- name: GetActiveSubscriptionsCount :one
SELECT COUNT(*) as count
FROM users_subscriptions
WHERE status IN ('active', 'trialing');

-- name: GetSubscriptionsByTier :many
SELECT 
    st.name as tier_name,
    COUNT(*) as count,
    SUM(st.price_cents) as total_revenue_cents
FROM users_subscriptions us
JOIN subscription_tiers st ON us.tier_id = st.id
WHERE us.status IN ('active', 'trialing')
GROUP BY st.name, st.price_cents
ORDER BY st.price_cents DESC;

-- name: GetMonthlyRecurringRevenue :one
SELECT COALESCE(SUM(st.price_cents), 0) as mrr_cents
FROM users_subscriptions us
JOIN subscription_tiers st ON us.tier_id = st.id
WHERE us.status IN ('active', 'trialing');

-- name: GetChurnedSubscriptionsThisMonth :one
SELECT COUNT(*) as count
FROM users_subscriptions
WHERE status = 'canceled'
AND updated_at >= date_trunc('month', NOW());

-- name: GetNewSubscriptionsThisMonth :one
SELECT COUNT(*) as count
FROM users_subscriptions
WHERE status IN ('active', 'trialing')
AND created_at >= date_trunc('month', NOW());

-- name: GetRevenueByMonth :many
SELECT 
    date_trunc('month', created_at) as month,
    SUM(amount_cents) as revenue_cents,
    COUNT(*) as payment_count
FROM payments
WHERE status = 'succeeded'
AND created_at >= $1
GROUP BY date_trunc('month', created_at)
ORDER BY month DESC;

-- name: GetTopPayingUsers :many
SELECT 
    u.id,
    u.name,
    SUM(p.amount_cents) as total_paid_cents,
    COUNT(p.id) as payment_count
FROM users u
JOIN payments p ON u.id = p.user_id
WHERE p.status = 'succeeded'
GROUP BY u.id, u.name
ORDER BY total_paid_cents DESC
LIMIT $1;

-- ==========================================
-- TRIAL MANAGEMENT
-- ==========================================

-- name: StartTrial :one
INSERT INTO users_subscriptions (
    id,
    user_id,
    tier_id,
    status,
    current_period_start,
    current_period_end,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    'trialing',
    NOW(),
    NOW() + interval '14 days',
    NOW(),
    NOW()
)
RETURNING *;

-- name: GetExpiringTrials :many
SELECT us.*, u.name as user_name, u.email as user_email
FROM users_subscriptions us
JOIN users u ON us.user_id = u.id
WHERE us.status = 'trialing'
AND us.current_period_end <= NOW() + interval '3 days'
AND us.current_period_end > NOW();

-- name: ConvertTrialToActive :exec
UPDATE users_subscriptions
SET 
    status = 'active',
    stripe_subscription_id = $2,
    updated_at = $3
WHERE id = $1;

-- ==========================================
-- DOWNGRADE/UPGRADE MANAGEMENT
-- ==========================================

-- name: ScheduleDowngrade :exec
UPDATE users_subscriptions
SET 
    tier_id = $2,
    cancel_at_period_end = false,
    updated_at = $3
WHERE id = $1;

-- name: GetPendingDowngrades :many
SELECT us.*, st.name as new_tier_name
FROM users_subscriptions us
JOIN subscription_tiers st ON us.tier_id = st.id
WHERE us.current_period_end <= NOW()
AND us.cancel_at_period_end = false;

-- ==========================================
-- FAILED PAYMENT HANDLING
-- ==========================================

-- name: GetFailedPayments :many
SELECT p.*, u.name as user_name, u.email as user_email
FROM payments p
JOIN users u ON p.user_id = u.id
WHERE p.status = 'failed'
AND p.created_at >= $1
ORDER BY p.created_at DESC;

-- name: MarkSubscriptionPastDue :exec
UPDATE users_subscriptions
SET 
    status = 'past_due',
    updated_at = $2
WHERE id = $1;

-- name: GetPastDueSubscriptions :many
SELECT us.*, u.name as user_name, u.email as user_email
FROM users_subscriptions us
JOIN users u ON us.user_id = u.id
WHERE us.status = 'past_due';

-- ==========================================
-- REFUNDS
-- ==========================================

-- name: CreateRefund :exec
UPDATE payments
SET status = 'refunded'
WHERE id = $1;

-- name: GetRefundedPayments :many
SELECT p.*, u.name as user_name
FROM payments p
JOIN users u ON p.user_id = u.id
WHERE p.status = 'refunded'
AND p.created_at >= $1
ORDER BY p.created_at DESC;