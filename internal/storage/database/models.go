// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0

package db

import ()

type InitialMigration struct {
	Message string
}

type ScratchSession struct {
	ID           int32
	UserID       int32
	RefreshToken string
	LoginDate    string
}

type ScratchUser struct {
	ID       int32
	Name     string
	Email    string
	Password string
}
