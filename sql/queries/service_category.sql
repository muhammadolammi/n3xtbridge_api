-- name: GetActiveServiceCategories :many
SELECT
    sc.id,
    sc.slug,
    sc.name,
    sc.description,
    sc.icon,
    COUNT(s.id)::int AS service_count
FROM service_categories sc
LEFT JOIN services s
    ON s.category_id = sc.id
WHERE sc.status = 'active'
GROUP BY sc.id
ORDER BY sc.created_at DESC;
-- name: GetServiceCategories :many
SELECT
    sc.id,
    sc.slug,
    sc.name,
    sc.description,
    sc.icon,
    COUNT(s.id) AS service_count
FROM service_categories sc
LEFT JOIN services s
    ON s.category_id = sc.id
GROUP BY sc.id
ORDER BY sc.created_at DESC;

-- name: CreateServiceCategory :one
INSERT INTO service_categories (
    slug, name, description, icon
) VALUES (
    $1, $2, $3, $4
) RETURNING *;