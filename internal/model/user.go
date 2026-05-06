package model

import (
	"errors"
	"time"
)

var ErrUserExists = errors.New("user already exists")

type User struct {
	ID        int64     `db:"id"`
	Login     string    `db:"login"`
	Password  string    `db:"password_hash,omitempty"`
	CreatedAt time.Time `db:"created_at"`
}

type UserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
