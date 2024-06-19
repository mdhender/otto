-- name: FetchSchemaMetadata :many
SELECT version, assets_path, templates_path
FROM metadata;

-- name: UpdateOttoMagic :exec
UPDATE users
SET magic = ?1
WHERE id = 1
  AND handle = 'otto'
  AND clan = '0000';

-- name: UpdateOttoPassword :exec
UPDATE users
SET hashed_password = ?1
WHERE id = 1
  AND handle = 'otto'
  AND clan = '0000';

-- name: CreateUser :one
INSERT INTO users (handle, hashed_password, clan, magic, enabled)
VALUES (?1, ?2, ?3, ?4, ?5)
RETURNING id;

-- name: DeleteUser :exec
UPDATE users
SET enabled = 'N'
WHERE id = ?1;

-- name: FetchUser :one
SELECT id, hashed_password, clan, magic, enabled
FROM users
WHERE handle = ?1;

-- name: UpdateUserMagic :exec
UPDATE users
SET magic = ?2
WHERE id = ?1
  AND id != 1;

-- name: UpdateUserPassword :exec
UPDATE users
SET hashed_password = ?2
WHERE id = ?1
  AND id != 1;
