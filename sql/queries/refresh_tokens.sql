-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, expires_at, user_id)
VALUES ( $1, NOW(), NOW(), $2, $3)
RETURNING *;

-- name: DeleteRefreshTokens :exec
DELETE FROM refresh_tokens;

-- name: ReadRefreshToken :one
SELECT token, user_id, created_at, updated_at, expires_at, revoked_at
FROM refresh_tokens
WHERE token = $1;

-- name: GetUserFromRefreshToken :one
SELECT users.id, users.email
FROM users
JOIN refresh_tokens
ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token = $1;
