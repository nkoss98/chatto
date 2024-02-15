// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: account.sql

package db

import (
	"context"
)

const createSession = `-- name: CreateSession :exec
INSERT INTO scratch.session (user_id, refresh_token, login_date) VALUES ($1, $2, $3)
`

type CreateSessionParams struct {
	UserID       int32
	RefreshToken string
	LoginDate    string
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) error {
	_, err := q.db.ExecContext(ctx, createSession, arg.UserID, arg.RefreshToken, arg.LoginDate)
	return err
}

const createUser = `-- name: CreateUser :one
INSERT INTO scratch.user (name, email, password)
VALUES ($1, $2, $3)
RETURNING id, name, email, password
`

type CreateUserParams struct {
	Name     string
	Email    string
	Password string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (ScratchUser, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.Name, arg.Email, arg.Password)
	var i ScratchUser
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Password,
	)
	return i, err
}

const getSession = `-- name: GetSession :one
SELECT id, user_id, refresh_token, login_date FROM scratch.session WHERE refresh_token = $1 AND user_id = $2
`

type GetSessionParams struct {
	RefreshToken string
	UserID       int32
}

func (q *Queries) GetSession(ctx context.Context, arg GetSessionParams) (ScratchSession, error) {
	row := q.db.QueryRowContext(ctx, getSession, arg.RefreshToken, arg.UserID)
	var i ScratchSession
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.RefreshToken,
		&i.LoginDate,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, name, email, password FROM scratch.user WHERE email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (ScratchUser, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i ScratchUser
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Password,
	)
	return i, err
}