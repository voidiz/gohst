package server

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// CreateAccount creates an account using the supplied username,
// then returns a randomly generated password.
func CreateAccount(newUser string) {
	db := Initialize()

	var user User
	err := db.QueryRowx("SELECT * FROM users WHERE username=?", newUser).
		StructScan(&user)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	if user.Username != "" {
		fmt.Println("Username already taken!")
		os.Exit(1)
	}

	pass, err := generatePassword()
	if err != nil {
		panic(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	db.MustExec("INSERT INTO users (username, password) VALUES (?, ?)", newUser, hash)
	fmt.Printf("Created user \"%s\".\nPassword:\n%s\n", newUser, pass)
}

// RegeneratePassword creates a new password for the supplied username
func RegeneratePassword(user string) {
	db := Initialize()

	var usercheck User
	err := db.QueryRowx("SELECT * FROM users WHERE username=?", user).
		StructScan(&usercheck)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("User \"%s\" does not exist!\n", user)
			os.Exit(1)
		}
		panic(err)
	}

	pass, err := generatePassword()
	if err != nil {
		panic(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("UPDATE users SET password=? WHERE username=?", hash, user)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Successfully changed password for \"%s\"!\nPassword:\n%s\n",
		user, pass)
}

// DeleteAccount deletes an account using the supplied username
func DeleteAccount(user string) {
	db := Initialize()

	var usercheck User
	err := db.QueryRowx("SELECT * FROM users WHERE username=?", user).
		StructScan(&usercheck)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("User \"%s\" does not exist!\n", user)
			os.Exit(1)
		}
		panic(err)
	}

	_, err = db.Exec("DELETE FROM users WHERE username=?", user)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Successfully deleted user \"%s\"!\n", user)
}

func generatePassword() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
