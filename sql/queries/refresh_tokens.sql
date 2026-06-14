-- name: CreateRefreshToken :one
insert into refresh_tokens (token, created_at, updated_at, user_id, expires_at)
values ($1, now(), now(), $2, now() + interval '60 days')
returning *;

-- name: RevokeRefreshToken :exec
update refresh_tokens
set updated_at = now(),
revoked_at = now()
where token = $1;

-- name: GetUserFromRefreshToken :one
select users.* from users
join refresh_tokens on refresh_tokens.user_id = users.id
where refresh_tokens.token = $1
and refresh_tokens.revoked_at is null
and refresh_tokens.expires_at > now();
