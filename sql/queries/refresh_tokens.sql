-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
user_id, expires_at,
token  )
VALUES ( $1, $2, $3)
RETURNING *;


-- name: UpdateRefreshToken :exec
UPDATE refresh_tokens
SET 
  token = $1,
  replaced_by = $2,
  revoked=$3
WHERE id = $4;


-- name: RefreshTokenExists :one
SELECT EXISTS (
    SELECT 1
    FROM refresh_tokens
    WHERE token = $1
);
-- name: GetRefreshToken :one
SELECT *
FROM refresh_tokens
WHERE token = $1;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token=$1;



-- name: RevokeRefreshTokens :exec
UPDATE refresh_tokens
SET 
  revoked=1
WHERE user_id = $1;