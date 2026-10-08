-- name: CreateOrganizationQuery :one
INSERT INTO organizations (name, domain, slug, plan, owner_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: GetOrganizationQuery :one
SELECT id, name, slug, domain, plan, created_at, updated_at
FROM organizations
WHERE id = $1;

-- name: UpdateOrganizationQuery :exec
UPDATE organizations
SET
    name = $2,
    domain = $3,
    slug = $4,
    updated_at = now()
WHERE id = $1;

-- name: DeleteOrganizationQuery :exec
DELETE
FROM organizations
WHERE id = $1;

-- name: GetDefaultOrganizationQuery :one
SELECT
    o.id,
    o.name,
    o.slug,
    o.domain,
    o.plan,
    o.created_at,
    o.updated_at
FROM organizations o
WHERE o.owner_id = $1
   OR EXISTS (
    SELECT 1
    FROM public.members m
    WHERE m.organization_id = o.id
      AND m.user_id = $1
)
ORDER BY o.created_at ASC, o.id ASC
LIMIT 1;

