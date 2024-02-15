-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA scratch;

CREATE TABLE scratch.user (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL
);

CREATE TABLE scratch.session (
    id SERIAL PRIMARY KEY,
    user_id integer NOT NULL,
    refresh_token VARCHAR(1000) NOT NULL,
    login_date VARCHAR(255) NOT NULL,

    CONSTRAINT fk_session_user FOREIGN KEY (user_id)
    REFERENCES scratch.user (id)
    ON DELETE SET NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS scratch.session;

DROP TABLE IF EXISTS scratch.user;
-- +goose StatementEnd

