-- +goose Up
-- +goose StatementBegin
CREATE TABLE initial_migration(
    message text NOT NULL
);

INSERT
INTO initial_migration (message)
VALUES ('successful migration');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE initial_migration;
-- +goose StatementEnd
