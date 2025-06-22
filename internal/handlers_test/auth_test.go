package handlers_test

import (
	"database/sql"
	"forum/internal/handlers"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE,
			username TEXT UNIQUE,
			password TEXT
		);
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER,
			expires_at TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestRegister_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tmpl := template.New("").Funcs(template.FuncMap{
		"inSlice": func(slice []string, val string) bool {
			for _, s := range slice {
				if s == val {
					return true
				}
			}
			return false
		},
	})
	tmpl = template.Must(tmpl.ParseGlob("../../templates/*.html"))

	handler := handlers.AuthHandler{
		DB:        db,
		Templates: tmpl,
		Err:       &handlers.ErrorHandler{Templates: tmpl},
	}

	form := url.Values{}
	form.Set("email", "test@example.com")
	form.Set("username", "testuser")
	form.Set("password", "123456")

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("ожидался редирект, получено: %d", resp.StatusCode)
	}
}

func TestLogin_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Создание пользователя вручную
	hashed, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	db.Exec(`INSERT INTO users (email, username, password) VALUES (?, ?, ?)`, "test@example.com", "testuser", string(hashed))

	tmpl := template.New("").Funcs(template.FuncMap{
		"inSlice": func(slice []string, val string) bool {
			for _, s := range slice {
				if s == val {
					return true
				}
			}
			return false
		},
	})
	tmpl = template.Must(tmpl.ParseGlob("../../templates/*.html"))

	handler := handlers.AuthHandler{DB: db, Templates: tmpl, Err: &handlers.ErrorHandler{Templates: tmpl}}

	form := url.Values{}
	form.Set("email", "test@example.com")
	form.Set("password", "123456")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("ожидался редирект после входа, получен статус %d", resp.StatusCode)
	}
}
