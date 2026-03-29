-- name: CreateQuote :one
INSERT INTO quotes (quote_request_id,user_id, amount, breakdown,discounts, notes, expires_at)
VALUES ($1, $2, $3, $4, $5,$6,$7)
RETURNING *; 

-- name: GetUserQuotesWithService :many
SELECT 
    q.*, 
    s.name as service_name,
    s.icon as service_icon,
        s.id as service_id

FROM quotes q 
JOIN quote_requests qr ON q.quote_request_id = qr.id
JOIN services s ON qr.service_id = s.id
WHERE qr.user_id = $1 AND  q.status IN ('sent', 'accepted', 'declined')
ORDER BY q.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUserQuoteWithService :one
SELECT 
    q.*, 
    s.name as service_name,
    s.icon as service_icon,
    s.id as service_id
FROM quotes q 
JOIN quote_requests qr ON q.quote_request_id = qr.id
JOIN services s ON qr.service_id = s.id
WHERE qr.user_id = $1 AND q.id= $2;

-- name: GetQuotes :many
SELECT    q.*, 
    s.name as service_name,
    s.icon as service_icon
FROM quotes q 
JOIN quote_requests qr ON q.quote_request_id = qr.id
JOIN services s ON qr.service_id = s.id
LIMIT $1 OFFSET $2;


-- name: CountQuotes :one
SELECT COUNT(*) FROM quotes;

-- name: CountUserQuotes :one
SELECT COUNT(q.id)
FROM quotes q
JOIN quote_requests qr ON q.quote_request_id = qr.id
WHERE qr.user_id = $1;

-- name: UpdateQuoteStatus :exec
UPDATE quotes SET status = $2, updated_at = NOW() WHERE id = $1;


-- name: GetQuote :one
SELECT * FROM quotes
WHERE id=$1;


-- name: GetInvoiceByQuoteID :one
SELECT * FROM invoices WHERE quote_id = $1;
