-- name: CreateQuoteRequest :one
INSERT INTO quote_requests (user_id, service_id,  description, attachments, promo_ids, vn_r2_key, video_key)
VALUES ($1, $2, $3, $4,$5, $6, $7)
RETURNING *;

-- name: GetQuoteRequests :many
SELECT 
    qr.*, 
    u.email as user_email, 
    u.first_name as user_name,
    s.name as service_name
FROM quote_requests qr
JOIN users u ON qr.user_id = u.id
JOIN services s ON qr.service_id = s.id
ORDER BY qr.created_at DESC
LIMIT $1 OFFSET $2;
-- name: GetUserQuoteRequests :many
SELECT 
  qr.*,
  s.name AS service_name,
  q.id AS quote_id
FROM quote_requests qr
JOIN services s ON qr.service_id = s.id
-- Use LEFT JOIN so requests without quotes still show up
LEFT JOIN quotes q ON q.quote_request_id = qr.id 
WHERE qr.user_id = $1
ORDER BY qr.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountQuoteRequests :one
SELECT COUNT(*) FROM quote_requests;

-- name: CountUserQuoteRequests :one
SELECT COUNT(*) FROM quote_requests WHERE user_id=$1;


-- name: UpdateQuoteRequestStatus :exec
UPDATE quote_requests SET status = $2, updated_at = NOW() WHERE id = $1;

-- name: UpdateQuoteRequestDescription :exec
UPDATE quote_requests SET description = $2, updated_at = NOW() WHERE id = $1;

-- name: GetQuoteRequest :one
SELECT * FROM quote_requests
WHERE id=$1;