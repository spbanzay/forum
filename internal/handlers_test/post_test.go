package handlers_test

import (
	"forum/internal/handlers"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestCreatePost_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Таблицы и пользователь
	db.Exec(`CREATE TABLE posts (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, title TEXT, content TEXT, created_at TIMESTAMP);`)
	db.Exec(`CREATE TABLE categories (id INTEGER PRIMARY KEY, name TEXT);`)
	db.Exec(`CREATE TABLE post_categories (post_id INTEGER, category_id INTEGER);`)
	db.Exec(`INSERT INTO users (id, email, username, password) VALUES (1, 'user@example.com', 'user1', 'pass')`)
	db.Exec(`INSERT INTO sessions (id, user_id, expires_at) VALUES ('session123', 1, datetime('now', '+1 hour'))`)
	db.Exec(`INSERT INTO categories (id, name) VALUES (1, 'Go'), (2, 'Web')`)

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

	handler := handlers.PostHandler{DB: db, Templates: tmpl, Err: &handlers.ErrorHandler{Templates: tmpl}}

	form := url.Values{}
	form.Set("title", "Test Post")
	form.Set("content", "This is a post.")
	form.Add("categories", "1")

	req := httptest.NewRequest(http.MethodPost, "/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "session123"})

	w := httptest.NewRecorder()
	handler.CreatePost(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected redirect, got %d", w.Code)
	}
}
