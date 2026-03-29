-- name: CreatePayment :one
INSERT INTO payments (
    invoice_id, amount, reference, status
) VALUES (
    $1, $2, $3, $4
) RETURNING *;




-- name: GetPaymentByReference :one
SELECT * FROM payments WHERE reference = $1;




-- name: UpdatePaymentStatus :exec
UPDATE payments SET
    status = $2,
    external_id = $3
WHERE reference = $1;

-- name: GetLatestPendingPayment :one
SELECT * FROM payments
WHERE invoice_id = $1 
  AND status = 'pending'
ORDER BY created_at DESC
LIMIT 1;