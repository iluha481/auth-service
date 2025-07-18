CREATE TABLE IF NOT EXISTS users (
    id           BIGSERIAL PRIMARY KEY,
    email        VARCHAR(255) NOT NULL UNIQUE,
    pass_hash    BYTEA NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

CREATE TABLE IF NOT EXISTS apps (
    id     BIGSERIAL PRIMARY KEY,
    name   VARCHAR(255) NOT NULL UNIQUE,
    secret VARCHAR(255) NOT NULL UNIQUE,
    refresh_secret VARCHAR(255) NOT NULL UNIQUE
);