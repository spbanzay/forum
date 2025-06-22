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

func TestAddComment_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Exec(`CREATE TABLE comments (id INTEGER PRIMARY KEY, post_id INTEGER, user_id INTEGER, content TEXT, created_at TIMESTAMP);`)
	db.Exec(`INSERT INTO users (id, email, username, password) VALUES (1, 'user@example.com', 'user1', 'pass')`)
	db.Exec(`INSERT INTO posts (id, user_id, title, content, created_at) VALUES (1, 1, 'Title', 'Body', datetime('now'))`)
	db.Exec(`INSERT INTO sessions (id, user_id, expires_at) VALUES ('session123', 1, datetime('now', '+1 hour'))`)

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

	handler := handlers.CommentHandler{DB: db, Templates: tmpl, Err: &handlers.ErrorHandler{Templates: tmpl}}

	form := url.Values{}
	form.Set("post_id", "1")
	form.Set("content", "Nice post!")

	req := httptest.NewRequest(http.MethodPost, "/post/comment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "session123"})

	w := httptest.NewRecorder()
	handler.AddComment(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected redirect, got %d", w.Code)
	}
}
