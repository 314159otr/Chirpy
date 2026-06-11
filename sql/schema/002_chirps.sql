-- +goose Up
create table chirps(
	id uuid primary key,
	user_id uuid not null references users(id) on delete cascade,
	created_at timestamp not null,
	updated_at timestamp not null,
	body text not null
);
-- +goose Down
drop table chirps;
