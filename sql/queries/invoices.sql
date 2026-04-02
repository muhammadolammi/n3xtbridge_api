-- name: CreateInvoice :one
INSERT INTO invoices (
    invoice_number, customer_name, customer_email, customer_phone, total, notes , items , discounts, user_id, payment_token
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,$8,$9, $10
) RETURNING *;

-- name: CreateInvoiceWithQuote :one
INSERT INTO invoices (
    invoice_number, customer_name, customer_email, customer_phone, total, notes , items , discounts, user_id,quote_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,$8,$9,$10
) RETURNING *;
-- name: GetInvoiceByNumber :one
SELECT * FROM invoices WHERE invoice_number = $1;

-- name: GetInvoice :one
SELECT * FROM invoices WHERE id = $1;

-- name: GetWorkersCreatedInvoices :many
SELECT * FROM invoices WHERE user_id = $1
ORDER BY created_at DESC 
LIMIT $2
OFFSET $3 ;

-- name: GetCustomerInvoices :many
SELECT * FROM invoices WHERE customer_email = $1
AND deleted_at IS NULL
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

-- -- name: DeleteInvoice :exec
-- DELETE FROM invoices WHERE id = $1;

-- name: CountInvoices :one
SELECT COUNT(*) FROM invoices;

-- name: CountWorkersCreatedInvoices :one
SELECT COUNT(*) FROM invoices WHERE user_id=$1;

-- name: CountCustomersInvoices :one
SELECT COUNT(*) FROM invoices WHERE customer_email=$1;


-- name: SoftDeleteInvoice :exec
UPDATE invoices 
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND customer_email = $2;

-- name: MarkInvoiceAsPaid :exec
UPDATE invoices 
SET status = 'paid', updated_at = NOW()
WHERE id = $1;


-- name: UpdateInvoiceReminderSentAt :exec
UPDATE invoices
SET reminder_sent_at = NOW(), updated_at = NOW()
WHERE id = $1;