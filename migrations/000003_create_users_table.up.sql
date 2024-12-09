CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    username TEXT NOT NULL,
    email citext UNIQUE NOT NULL,
    password bytea NOT NULL,
    activated bool NOT NULL,
    version INT NOT NULL DEFAULT 1
);