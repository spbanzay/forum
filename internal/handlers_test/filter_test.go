package handlers_test

import (
	"forum/internal/handlers"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestFilteredPosts_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Exec(`CREATE TABLE posts (id INTEGER PRIMARY KEY, user_id INTEGER, title TEXT, content TEXT, created_at TIMESTAMP);`)
	db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT, username TEXT, password TEXT);`)
	db.Exec(`CREATE TABLE categories (id INTEGER PRIMARY KEY, name TEXT);`)
	db.Exec(`CREATE TABLE post_categories (post_id INTEGER, category_id INTEGER);`)
	db.Exec(`CREATE TABLE post_likes (post_id INTEGER, user_id INTEGER, is_like BOOLEAN);`)

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

	handler := handlers.FilterHandler{DB: db, Templates: tmpl, Err: &handlers.ErrorHandler{Templates: tmpl}}

	req := httptest.NewRequest(http.MethodGet, "/?q=hello", nil)
	w := httptest.NewRecorder()
	handler.FilteredPosts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
}
