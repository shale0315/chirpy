-- name: CreateUser :one

INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;


-- name: ResetUsers :execresult

DELETE FROM users;


-- name: Login :one

SELECT * FROM users WHERE email=$1;


-- name: StoreRefreshToken :one

INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    NULL
)
RETURNING *;

-- name: GetUserFromRefresh :one

SELECT user_id,expires_at,revoked_at FROM refresh_tokens WHERE token=$1;

-- name: RevokeToken :execresult

UPDATE refresh_tokens
SET revoked_at=NOW(), updated_at=NOW()
WHERE token=$1;
