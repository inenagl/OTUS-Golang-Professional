-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS Events (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    title VARCHAR(255),
    description TEXT,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    user_id UUID,
    notify_before BIGINT
);
CREATE INDEX IF NOT EXISTS user_idx ON Events (user_id);
CREATE INDEX IF NOT EXISTS start_idx ON Events (start_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS user_idx;
DROP INDEX IF EXISTS start_idx;
DROP TABLE IF EXISTS Events;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd
