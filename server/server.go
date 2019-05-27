package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

// Server defines the database connection and the HTTP server
// of the application
type Server struct {
	DB     *sqlx.DB
	Router *chi.Mux
}

// Initialize creates the database connection
func Initialize() *sqlx.DB {
	db, err := sqlx.Connect("mysql", fmt.Sprintf("%v:%v@/gohst?parseTime=true",
		viper.GetString("dbUser"), viper.GetString("dbPass")))
	if err != nil {
		panic(err)
	}

	return db
}

// TODO: add mode for development and production
// Run starts the server
func (s *Server) Run(mode string) {
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

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", viper.GetInt("port")), s.Router))
}
