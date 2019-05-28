package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// CreateConfig creates the server configuration file
func CreateConfig() {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	path := filepath.Join(home, ".gohst.yaml")
	content := []byte(configText)

	if err := ioutil.WriteFile(path, content, 0644); err != nil {
		panic(err)
	}
	fmt.Printf("Created configuration file %s!\nDon't forget to fill in your database credentials!\n", path)
}

// Setup runs the initial setup using the supplied settings in the configuration file
func Setup() {
	db, err := sqlx.Connect("mysql", fmt.Sprintf("%v:%v@tcp(127.0.0.1:3306)/?multiStatements=true",
		viper.GetString("dbUser"), viper.GetString("dbPass")))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE DATABASE gohst")
	if me, ok := err.(*mysql.MySQLError); ok {
		if me.Number == 1007 {
			if confirmAction("Database already exists, do you want to delete it?") {
				db.MustExec("DROP DATABASE gohst; CREATE DATABASE gohst;")
			} else {
				fmt.Println("Aborting...")
				os.Exit(1)
			}
		} else {
			panic(err)
		}
	} else if !ok {
		panic(err)
	}

	db.MustExec(dbStructure)
	fmt.Println("Successfully setup the database!")

	if err := os.Mkdir(viper.GetString("staticDir"), 0755); err != nil {
		panic(err)
	} else {
		fmt.Println("Created static file directory!")
	}

	defer db.Close()
	fmt.Println("Done with setup!")
}

func confirmAction(action string) bool {
	var s string

	fmt.Printf("%s (y/N): ", action)
	if _, err := fmt.Scan(&s); err != nil {
		panic(err)
	}

	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if s == "yes" || s == "y" {
		return true
	}
	return false
}

var configText = `
## uncomment the following lines and fill in the
## necessary information

######################
## db configuration ##
######################
# dbUser:
# dbPass:

##########################
## server configuration ##
##########################
# domain: e.g. mywebsite.com 
# staticDir: /path/to/static/dir (defaults to "static" in the current directory)
# port: defaults to 80
# maxFileSize: bytes, defaults to 5000000 (5 MB)
# blockedMimeTypes: 
# - list of blocked mime types
# - defaults to 
# - application/x-dosexec
# - application/x-executable`

var dbStructure = `
USE gohst;

CREATE TABLE users (
	id int(11) NOT NULL AUTO_INCREMENT,
	username varchar(255) NOT NULL,
	password varchar(255) NOT NULL,
	created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY(id)
) ENGINE=InnoDB;

CREATE TABLE auth_tokens (
	id int(11) NOT NULL AUTO_INCREMENT,
	account_id int(11) NOT NULL,
	token varchar(255) NOT NULL,
	created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

	PRIMARY KEY(id),
	INDEX acc_ind (account_id),
	FOREIGN KEY (account_id)
		REFERENCES users(id)
		ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE user_files (
	id int(11) NOT NULL AUTO_INCREMENT,
	account_id int(11) NOT NULL,
	name varchar(255) NOT NULL,
	created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

	PRIMARY KEY(id),
	INDEX acc_ind (account_id),
	FOREIGN KEY (account_id)
		REFERENCES users(id)
		ON DELETE CASCADE
) ENGINE=InnoDB;`
