-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN email VARCHAR(255) UNIQUE NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN email;
-- +goose StatementEnd
