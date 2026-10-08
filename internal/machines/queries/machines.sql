-- name: RegisterMachineQuery :one
INSERT INTO machines (organization_id, name)
VALUES ($1, $2)
RETURNING id;

-- name: RegisterTokenMachineQuery :exec
INSERT INTO machine_enrollment_tokens (machine_id, token_hash, expires_at, created_by_user_id)
VALUES ($1, $2, $3, $4);

-- name: GetMachineQuery :one
SELECT
    id,
    organization_id,
    name,
    hostname,
    agent_version,
    last_seen_at,
    created_at,
    updated_at
FROM machines
WHERE id = $1;

-- name: ConsumeMachineTokenQuery :one
UPDATE machine_enrollment_tokens
SET
    used_at = now(),
    updated_at = now()
WHERE token_hash = $1
  AND machine_id = $2
  AND used_at IS NULL
  AND expires_at > now()
RETURNING
    id,
    machine_id,
    token_hash,
    expires_at
;

-- name: ListMachineQuery :many
SELECT
    id,
    organization_id,
    name,
    hostname,
    agent_version,
    last_seen_at,
    created_at,
    updated_at
FROM machines
WHERE organization_id = $1
ORDER BY created_at;

-- name: UpdateMachineQuery :exec
UPDATE machines
SET
    hostname = $2,
    agent_version = $3,
    last_seen_at = $4,
    updated_at = now()
WHERE id = $1;

-- name is changed separately from agent heartbeat updates.
-- name: UpdateMachineNameQuery :exec
UPDATE machines
SET name = $2, updated_at = now()
WHERE id = $1;

-- name: DeleteMachineQuery :exec
DELETE
FROM machines
WHERE id = $1;
