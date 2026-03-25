-- name: CreateInvoice :one
INSERT INTO invoices (
    invoice_number, customer_name, customer_email, customer_phone, total, notes , items , discounts, user_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,$8, $9
) RETURNING *;


-- name: GetInvoiceByNumber :one
SELECT * FROM invoices WHERE invoice_number = $1;

-- name: GetInvoice :one
SELECT * FROM invoices WHERE id = $1;

-- name: GetUserInvoices :many
SELECT * FROM invoices WHERE user_id = $1
ORDER BY created_at DESC 
LIMIT $2
OFFSET $3 ;


-- name: ListInvoices :many
SELECT * FROM invoices ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: UpdateInvoice :one
UPDATE invoices SET
    customer_name = $2,
    customer_email = $3,
    customer_phone = $4,
    total = $5,
    notes = $6,
    items =$7,
    discounts=$8,
    updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: DeleteInvoice :exec
DELETE FROM invoices WHERE id = $1;