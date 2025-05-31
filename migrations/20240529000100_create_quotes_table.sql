-- +goose Up
CREATE TABLE IF NOT EXISTS quotes (
    id SERIAL PRIMARY KEY,
    author VARCHAR(255) NOT NULL,
    quote TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down

