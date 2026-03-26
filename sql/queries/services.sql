-- name: CreateService :one
INSERT INTO services (
    name, description, category, is_featured, icon , tags , image
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;


-- name: GetServiceByName :one
SELECT * FROM services WHERE name = $1;

-- name: GetService :one
SELECT * FROM services WHERE id = $1;

-- name: GetActiveServices :many
-- Used for the landing page
SELECT * FROM services 
WHERE is_active = true
ORDER BY 
    is_featured DESC, 
    created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetServices :many
-- Used for the admin dashboard (shows everything)
SELECT * FROM services 
ORDER BY 
    is_featured DESC, 
    created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateServiceStatus :exec
UPDATE services SET is_active = $2 WHERE id = $1;

-- name: CountActiveServices :one
SELECT COUNT(*) FROM services WHERE is_active = true;
-- name: CountServices :one
SELECT COUNT(*) FROM services;