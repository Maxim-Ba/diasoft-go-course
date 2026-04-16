-- +goose Up
-- +goose StatementBegin
ALTER TABLE events ADD COLUMN notified_at TIMESTAMP NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events DROP COLUMN notified_at;
-- +goose StatementEnd
