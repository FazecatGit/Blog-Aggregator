-- +goose Up
CREATE TABLE users(
    id UUID  PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    username TEXT NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE users;