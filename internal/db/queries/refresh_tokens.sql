-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at, user_agent, ip)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT *
FROM refresh_tokens
WHERE token_hash = $1;

-- name: GetRefreshTokenByID :one
SELECT *
FROM refresh_tokens
WHERE id = $1;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE id = $1
  AND revoked_at IS NULL
RETURNING *;

-- name: RotateRefreshToken :one
UPDATE refresh_tokens
SET revoked_at     = NOW(),
    replaced_by_id = $2
WHERE id = $1
  AND revoked_at IS NULL
RETURNING *;

-- name: RevokeAllUserRefreshTokens :execrows
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1
  AND revoked_at IS NULL;

-- name: DeleteExpiredRefreshTokens :execrows
DELETE FROM refresh_tokens
WHERE expires_at < NOW() - INTERVAL '7 days';
