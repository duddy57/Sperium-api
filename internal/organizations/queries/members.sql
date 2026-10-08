-- name: CreateMemberOwnerQuery :exec
INSERT INTO members (user_id, organization_id, role)
VALUES ($1, $2, $3);

-- name: ListMembersQuery :many
SELECT m.id as memberId, m.role, u.id, u.name, u.email, m.created_at, m.updated_at FROM members m
LEFT JOIN public.users u ON m.user_id = u.id
WHERE m.organization_id = $1;

-- name: GetMemberByIDQuery :one
SELECT m.id as memberID, m.role, u.id, u.name, u.email, m.created_at, m.updated_at
FROM members m
LEFT JOIN public.users u ON m.user_id = u.id
WHERE m.id = $1 AND m.organization_id = $1;

-- name: RemoveMemberQuery :exec
DELETE FROM members
WHERE id = $1;
