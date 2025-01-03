-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES ( gen_random_uuid(), NOW(), NOW(), $1, $2)
RETURNING *;

-- name: DeleteChirp :exec
DELETE FROM chirps;

-- name: ReadChirp :one
SELECT id, created_at, updated_at, body, user_id
FROM chirps
WHERE id = $1;

-- name: ReadAllChirps :many
SELECT id, created_at, updated_at, body, user_id
FROM chirps
WHERE ($1::uuid IS NULL OR $1::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR user_id = $1::uuid)
ORDER BY created_at;

-- name: DeleteSpecificChirp :exec
DELETE FROM chirps
WHERE user_id = $1 AND id = $2;
