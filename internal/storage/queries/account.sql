-- name: GetUserByEmail :one
SELECT * FROM scratch.user WHERE email = $1;

-- name: CreateUser :one
INSERT INTO scratch.user (name, email, password)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateSession :exec
INSERT INTO scratch.session (user_id, refresh_token, login_date) VALUES ($1, $2, $3);

-- name: GetSession :one
SELECT * FROM scratch.session WHERE refresh_token = $1 AND user_id = $2;

-- tutaj left join jakis zeby zajebac usera z sesja i essa