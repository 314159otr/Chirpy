-- name: CreateUser :one
insert into users (id, created_at, updated_at, email, hashed_password)
values (gen_random_uuid(), now(), now(), $1, $2)
returning *;

-- name: GetUserByEmail :one
select * from users
where email = $1;

-- name: UpdateUserPasswordAndEmail :one
update users
set updated_at = now(),
hashed_password = $1,
email = $2
where id = $3
returning *;

-- name: DeleteUsers :exec
delete from users;
