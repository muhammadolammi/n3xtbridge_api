-- name: CreateUser :one
INSERT INTO users (
    email, password_hash, first_name, last_name, phone_number, address, role
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING id, email, password_hash, first_name, last_name, phone_number, address, role, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, first_name, last_name, phone_number, address, role, created_at, updated_at FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, first_name, last_name, phone_number, address, role, created_at, updated_at FROM users WHERE id = $1;

-- name: UpdateUserRole :one
UPDATE users SET
    role = $2,
    updated_at = NOW()
WHERE id = $1 RETURNING id, email, password_hash, first_name, last_name, phone_number, address, role, created_at, updated_at;
