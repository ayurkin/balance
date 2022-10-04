-- +goose Up

CREATE TABLE IF NOT EXISTS balance.balance
(
    user_id bigserial PRIMARY KEY,
    value decimal(10, 2) NOT NULL CHECK (value >= 0)
);