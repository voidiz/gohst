package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

type Server struct {
	DB     *sqlx.DB
	Router *chi.Mux
}

// CreateConfig creates the server configuration file
func CreateConfig() {
	content := []byte(`
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
# address: e.g. https://website.com
# staticDir: /path/to/static/dir (defaults to "static" in the current directory)
# port: defaults to 80
# maxFileSize: bytes, defaults to 5000000 (5 MB)
# blockedMimeTypes: 
# - list of blocked mime types
# - defaults to 
# - application/x-dosexec
# - application/x-executable`)

	if err := ioutil.WriteFile("./config.yaml", content, 0644); err != nil {
		panic(err)
	}
	fmt.Println("Created configuration file config.yaml in the current directory!")
}

// Initialize creates the database connection and reads the config file
func Initialize() *sqlx.DB {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic("Couldn't load config file, have you run 'gohst config create'?")
	}
	viper.SetConfigType("yaml")

	// Initialize db connection
	db, err := sqlx.Connect("mysql", fmt.Sprintf("%v:%v@/gohst?parseTime=true",
		viper.GetString("dbUser"), viper.GetString("dbPass")))
	if err != nil {
		panic(err)
	}

	return db
}

// Setup runs the initial setup using the supplied settings in the configuration file
func Setup() {
	// Default configuration
	viper.SetDefault("staticDir", "static")

	db := Initialize()

	dbStructure := `
		USE gohst;

		CREATE TABLE users (
			id int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
			username varchar(255) NOT NULL,
			password varchar(255) NOT NULL
		) ENGINE=InnoDB;

		CREATE TABLE auth_tokens (
			id int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
			account_id int(11) NOT NULL,
			token varchar(255) NOT NULL,
			created_at datetime DEFAULT CURRENT_TIMESTAMP NOT NULL,

			INDEX acc_ind (account_id),
			FOREIGN KEY (account_id)
				REFERENCES users(id)
				ON DELETE CASCADE
		) ENGINE=InnoDB;

		CREATE TABLE user_files (
			id int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
			account_id int(11) NOT NULL,
			name varchar(255) NOT NULL,

			INDEX acc_ind (account_id),
			FOREIGN KEY (account_id)
				REFERENCES users(id)
				ON DELETE CASCADE
		) ENGINE=InnoDB;`

	_, err := db.Exec(dbStructure)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Successfully created the database!")
	}

	if err := os.Mkdir(viper.GetString("staticDir"), 0755); err != nil {
		panic(err)
	} else {
		fmt.Println("Created static file directory!")
	}

	defer db.Close()
	fmt.Println("Done with setup!")
}

// Run starts the server
func (s *Server) Run(mode string) {
	// Default configuration
	viper.SetDefault("port", 80)
	viper.SetDefault("staticDir", "static")
	viper.SetDefault("maxFileSize", int64(5000000))
	viper.SetDefault("blockedMimeTypes", []string{"application/x-dosexec", "application/x-executable"})

	// Open DB and config
	s.DB = Initialize()

	// Initialize router
	s.Router = chi.NewRouter()
	e := Env{
		DB:               s.DB,
		StaticDir:        viper.GetString("staticDir"),
		MaxFileSize:      viper.Get("maxFileSize").(int64),
		BlockedMimeTypes: viper.GetStringSlice("blockedMimeTypes"),
	}

	// Routes
	s.Router.Use(middleware.Timeout(30 * time.Second))
	s.Router.Use(middleware.StripSlashes)
	s.Router.Group(func(r chi.Router) {
		// Public routes
		r.Get("/", e.ShowIndex)
		r.Get("/{filename:\\w+.\\w+}", e.GetFile)
		r.Post("/login", e.CreateAuthToken)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(e.AuthMiddleware)
			r.Post("/", e.UploadFile)
			// r.Delete("/{filename:\\w+.\\w+}", e.DeleteFile)
		})
	})

	// Scanner to delete old files
	// go tools.StartScanner(e.StaticDir, "1s")

	http.ListenAndServe(fmt.Sprintf(":%v", viper.GetInt("port")), s.Router)
}
