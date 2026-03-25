-- name: CreateUser :one
INSERT INTO users (
    email, password_hash, first_name, last_name, phone_number, address, role, country, state
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,$8,$9
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUserRole :one
UPDATE users SET
    role = $2,
    updated_at = NOW()
WHERE id = $1 RETURNING *;


-- name: UpdatePassword :exec
UPDATE users
SET 
  password_hash = $1
WHERE email = $2;


-- name: UserExists :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE email = $1
);