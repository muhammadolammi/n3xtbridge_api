-- name: CreateInvoice :one
INSERT INTO invoices (
    invoice_number, customer_name, customer_email, customer_phone, discount, total, notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: CreateItem :one
INSERT INTO items (
    invoice_id, name, quantity, price
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetInvoiceByNumber :one
SELECT * FROM invoices WHERE invoice_number = $1;

-- name: GetInvoiceByID :one
SELECT * FROM invoices WHERE id = $1;

-- name: GetItemsByInvoiceID :many
SELECT * FROM items WHERE invoice_id = $1;

-- name: ListInvoices :many
SELECT * FROM invoices ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: UpdateInvoice :one
UPDATE invoices SET
    customer_name = $2,
    customer_email = $3,
    customer_phone = $4,
    discount = $5,
    total = $6,
    notes = $7,
    updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: DeleteInvoice :exec
DELETE FROM invoices WHERE id = $1;