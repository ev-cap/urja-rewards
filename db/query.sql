-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByPhone :one
SELECT * FROM users
WHERE phone = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (phone)
VALUES ($1)
RETURNING *;

-- name: GetPointsEventsByUser :many
SELECT * FROM points_events
WHERE user_id = $1
ORDER BY occurred_at DESC;

-- name: CreatePointsEvent :one
INSERT INTO points_events (user_id, event_type, ref_id, points, meta)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserPointsBalance :one
SELECT COALESCE(SUM(points), 0)::bigint as balance
FROM points_events
WHERE user_id = $1;

-- name: GetRewardsCatalog :many
SELECT * FROM rewards_catalog
WHERE active = true
ORDER BY cost ASC;

-- name: GetReward :one
SELECT * FROM rewards_catalog
WHERE id = $1 LIMIT 1;

-- name: GetRedemptionsByUser :many
SELECT * FROM redemptions
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetRedemption :one
SELECT * FROM redemptions
WHERE id = $1 LIMIT 1;

-- name: CreateRedemption :one
INSERT INTO redemptions (user_id, reward_id, points_spent)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateRedemptionStatus :one
UPDATE redemptions
SET status = $2
WHERE id = $1
RETURNING *;

-- name: GetPendingRedemptionsOlderThan :many
SELECT * FROM redemptions
WHERE status = 'PENDING' AND created_at < $1
ORDER BY created_at ASC;

-- Rules queries
-- name: CreateRule :one
INSERT INTO rules (id, name, description, config, active, created_by) 
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetRule :one
SELECT * FROM rules WHERE id = $1;

-- name: ListRules :many
SELECT * FROM rules ORDER BY created_at DESC;

-- name: UpdateRule :one
UPDATE rules SET name = $2, description = $3, config = $4, active = $5, updated_at = NOW() 
WHERE id = $1 RETURNING *;

-- name: DeleteRule :exec
DELETE FROM rules WHERE id = $1;

-- Segments queries
-- name: CreateSegment :one
INSERT INTO segments (id, name, description, criteria, active, created_by) 
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetSegment :one
SELECT * FROM segments WHERE id = $1;

-- name: ListSegments :many
SELECT * FROM segments ORDER BY created_at DESC;

-- name: UpdateSegment :one
UPDATE segments SET name = $2, description = $3, criteria = $4, active = $5 
WHERE id = $1 RETURNING *;

-- Enhanced rewards queries
-- name: ListRewards :many
SELECT * FROM rewards_catalog WHERE active = $1 ORDER BY created_at DESC;

-- name: CreateReward :one
INSERT INTO rewards_catalog (id, name, description, cost, segment, active, created_by) 
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: UpdateReward :one
UPDATE rewards_catalog SET name = $2, description = $3, cost = $4, segment = $5, active = $6 
WHERE id = $1 RETURNING *; 