-- +goose Up

CREATE TABLE IF NOT EXISTS balance.history
(
    id          bigserial PRIMARY KEY,
    from_id     bigserial,
    to_id       bigserial,
    value       decimal(10, 2),
    occurred_at timestamptz NOT NULL,
    description text
);