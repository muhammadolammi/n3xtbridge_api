-- name: CreateServiceOrder :one
INSERT INTO service_orders (
    order_number, email, full_name, business_name, phone, whatsapp_phone,
    company_size, referral_source, service_type, appliance_details,
    delivery_address, transport_fee, promo_applied, status, user_id, notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
) RETURNING *;

-- name: GetServiceOrderByID :one
SELECT * FROM service_orders WHERE id = $1;

-- name: GetServiceOrderByNumber :one
SELECT * FROM service_orders WHERE order_number = $1;

-- name: GetServiceOrdersByEmail :many
SELECT * FROM service_orders WHERE email = $1 ORDER BY created_at DESC;

-- name: GetServiceOrdersByUserID :many
SELECT * FROM service_orders WHERE user_id = $1 ORDER BY created_at DESC;

-- name: UpdateServiceOrderStatus :one
UPDATE service_orders
SET status = $2, updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: ListServiceOrders :many
SELECT * FROM service_orders
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountServiceOrders :one
SELECT COUNT(*) FROM service_orders;

-- name: CountServiceOrdersByStatus :many
SELECT status, COUNT(*) as count FROM service_orders GROUP BY status;

-- name: CountServiceOrdersByServiceType :many
SELECT service_type, COUNT(*) as count FROM service_orders GROUP BY service_type;
