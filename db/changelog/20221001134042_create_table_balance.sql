-- +goose Up

CREATE TABLE IF NOT EXISTS balance.balance
(
    user_id bigserial PRIMARY KEY,
    balance decimal(10, 2) NOT NULL CHECK (balance >= 0)
);