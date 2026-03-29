-- name: CreatePromotion :one
INSERT INTO promotions (
    code,
    name,
    description,
    breakdown,
    is_active,
    starts_at,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetActivePromoByCode :one
SELECT * FROM promotions 
WHERE code = $1 
  AND is_active = true 
  AND (expires_at IS NULL OR expires_at > NOW())
LIMIT 1;

-- name: GetPromotionByID :one
SELECT * FROM promotions WHERE id = $1 LIMIT 1;

-- name: ListActivePromotions :many
SELECT * FROM promotions 
WHERE is_active = true 
ORDER BY created_at DESC;

-- name: UpdateServicePromo :exec
UPDATE services SET active_promo_ids = $2 WHERE id = $1;

-- name: ListPromos :many
SELECT * FROM promotions ORDER BY created_at DESC LIMIT $1 OFFSET $2;


-- name: CountPromos :one
SELECT COUNT(*) FROM promotions;

-- name: GetActivePromos :many
-- Used for the landing page
SELECT * FROM promotions 
WHERE is_active = true
ORDER BY 
    created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountActivePromos :one
SELECT COUNT(*) FROM promotions WHERE is_active = true;