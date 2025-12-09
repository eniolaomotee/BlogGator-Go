-- +goose Up
ALTER TABLE users_subscriptions ADD COLUMN email TEXT;
