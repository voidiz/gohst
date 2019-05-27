package server

import "time"

type User struct {
	ID        int
	Username  string
	Password  string
	CreatedAt time.Time `db:"created_at"`
}

type AuthToken struct {
	ID        int
	AccountID int `db:"account_id"`
	Token     string
	CreatedAt time.Time `db:"created_at"`
}

type UserFile struct {
	ID        int
	AccountID int `db:"account_id"`
	Name      string
	CreatedAt time.Time `db:"created_at"`
}
