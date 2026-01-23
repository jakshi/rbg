-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- +goose Down
-- no-op: keep extension installed
-- other objects/migrations may depend on pgcrypto
