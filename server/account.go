package server

import "database/sql"

func CreateAccount(newUser string) error {
	db := Initialize()

	var user string
	err := db.QueryRowx("SELECT username FROM users WHERE username=?").
		StructScan(&user)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	return nil
}
