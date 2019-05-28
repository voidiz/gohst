package server

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"github.com/voidiz/gohst/tools"
	"golang.org/x/crypto/bcrypt"
)

type Env struct {
	DB               *sqlx.DB
	StaticDir        string
	MaxFileSize      int64
	BlockedMimeTypes []string
}

type contextKey string

const accountIDKey contextKey = "AccountID"

func (e *Env) ShowIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello!"))
}

func (e *Env) GetFile(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.Dir(e.StaticDir)).ServeHTTP(w, r)
	// 	http.ServeFile(w, r, fmt.Sprintf("static/%v", chi.URLParam(r, "filename")))
}

func (e *Env) UploadFile(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		switch err.Error() {
		case "http: no such file":
			http.Error(w, "No file uploaded", http.StatusOK)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer file.Close()

	if header.Size >= e.MaxFileSize {
		http.Error(w, "File too large!", http.StatusBadRequest)
		return
	}

	if e.fileBlocked(header.Header.Get("Content-Type")) {
		http.Error(w, "File not allowed!", http.StatusUnsupportedMediaType)
		return
	}

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileName, err := tools.GenerateFileName(e.DB, fileBytes, header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	f, err := os.Create(filepath.Join(e.StaticDir, fileName))
	defer f.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err = f.Write(fileBytes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	e.DB.MustExec("INSERT INTO user_files (account_id, name) VALUES (?, ?)",
		r.Context().Value(accountIDKey), fileName)

	var scheme string
	if r.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}

	w.WriteHeader(http.StatusOK)
	resp := fmt.Sprintf("%s://%s/%s", scheme, r.Host, fileName)
	w.Write([]byte(resp))
}

func (e *Env) DeleteFile(w http.ResponseWriter, r *http.Request) {
	var (
		userCheck string
		fileCheck string
		err       error
	)
	fileName := chi.URLParam(r, "filename")

	err = e.DB.Get(&userCheck, "SELECT username FROM users WHERE id=?",
		r.Context().Value(accountIDKey))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "You are not the owner of the file", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = e.DB.Get(&fileCheck, "SELECT name FROM user_files WHERE name=?", fileName)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid filename", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = e.DB.Exec("DELETE FROM user_files WHERE name=?",
		fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.Remove(filepath.Join(e.StaticDir, fileName)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted " + fileName))
}

func (e *Env) CreateAuthToken(w http.ResponseWriter, r *http.Request) {
	formUser := r.PostFormValue("user")
	if formUser == "" {
		http.Error(w, "Missing user value", http.StatusBadRequest)
		return
	}
	formPass := r.PostFormValue("pass")
	if formPass == "" {
		http.Error(w, "Missing pass value", http.StatusBadRequest)
		return
	}

	var user User
	err := e.DB.QueryRowx("SELECT * FROM users WHERE username=?", formUser).
		StructScan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Server error, try again", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password),
		[]byte(formPass)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(16)
	if err != nil {
		http.Error(w, "Server error, try again", http.StatusInternalServerError)
		return
	}

	query := "INSERT INTO auth_tokens (account_id, token) VALUES (?, ?)"
	e.DB.MustExec(query, user.ID, token)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token))
}

func (e *Env) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var au AuthToken

		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Missing authorization header", http.StatusForbidden)
			return
		}
		token = strings.TrimLeft(token, "Bearer ")

		err := e.DB.QueryRowx("SELECT * FROM auth_tokens WHERE token=?", token).StructScan(&au)
		if err != nil {
			http.Error(w, "Invalid bearer token", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), accountIDKey, au.AccountID)

		// Next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateToken generates a bearer token for the authentication system
func generateToken(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (e *Env) fileBlocked(mimeType string) bool {
	for _, v := range e.BlockedMimeTypes {
		if v == mimeType {
			return true
		}
	}
	return false
}
