package server

import (
	"github.com/jmoiron/sqlx"
)

func CreateAccount(newUser string) error {
	db := Initialize()

	var user string
	err := db.QueryRowx("SELECT username FROM users WHERE username=?")
	if err != nil && err != sql.ErrNoRows {
		return err
	}
}
