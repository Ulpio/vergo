-- name: UpsertSubscription :exec
INSERT INTO subscriptions (org_id, stripe_customer_id, stripe_subscription_id, status, plan, current_period_end)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (org_id) DO UPDATE SET
    stripe_customer_id = EXCLUDED.stripe_customer_id,
    stripe_subscription_id = EXCLUDED.stripe_subscription_id,
    status = EXCLUDED.status,
    plan = EXCLUDED.plan,
    current_period_end = EXCLUDED.current_period_end,
    updated_at = now();

-- name: GetSubscriptionByOrg :one
SELECT id, org_id, stripe_customer_id, stripe_subscription_id, status, plan, current_period_end, created_at, updated_at
FROM subscriptions
WHERE org_id = $1;

-- name: GetSubscriptionByStripeCustomer :one
SELECT id, org_id, stripe_customer_id, stripe_subscription_id, status, plan, current_period_end, created_at, updated_at
FROM subscriptions
WHERE stripe_customer_id = $1;

-- name: UpdateSubscriptionStatus :exec
UPDATE subscriptions
SET status = $2, plan = $3, current_period_end = $4, updated_at = now()
WHERE stripe_subscription_id = $1;

-- name: CancelSubscription :exec
UPDATE subscriptions
SET status = 'canceled', updated_at = now()
WHERE stripe_subscription_id = $1;
