-- +goose Up
-- +goose StatementBegin
ALTER TABLE Events
ADD COLUMN notified_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS start_date_idx ON Events (start_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE Events
DROP COLUMN notified_at;
DROP INDEX IF EXISTS start_date_idx;
-- +goose StatementEnd
