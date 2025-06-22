package main

import (
	"database/sql"
	"fmt"
	dbinit "forum/internal/db"
	"forum/internal/handlers"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var (
	templates *template.Template
	db        *sql.DB
)

func main() {
	var err error

	db, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = dbinit.InitDatabase(db); err != nil {
		log.Fatal("Ошибка при инициализации схемы:", err)
	}

	templates = template.New("").Funcs(template.FuncMap{
		"inSlice": func(slice []string, val string) bool {
			for _, s := range slice {
				if s == val {
					return true
				}
			}
			return false
		},
		"contains": func(m map[string][]string, key string, val interface{}) bool {
			for _, v := range m[key] {
				if fmt.Sprint(v) == fmt.Sprint(val) {
					return true
				}
			}
			return false
		},
	})

	templates, err = templates.ParseGlob(filepath.Join("templates", "*.html"))
	if err != nil {
		log.Fatal("Ошибка парсинга шаблонов:", err)
	}

	for _, tmpl := range templates.Templates() {
		log.Println("Загружен шаблон:", tmpl.Name())
	}

	errHandler := &handlers.ErrorHandler{Templates: templates}

	commentHandler := handlers.CommentHandler{
		DB:        db,
		Templates: templates,
		Err:       errHandler,
	}

	likeHandler := handlers.LikeHandler{
		DB:  db,
		Err: errHandler,
	}

	filterHandler := handlers.FilterHandler{
		DB:        db,
		Templates: templates,
		Err:       errHandler,
	}

	postHandler := handlers.PostHandler{
		DB:        db,
		Templates: templates,
		Err:       errHandler,
	}

	authHandler := handlers.AuthHandler{
		DB:        db,
		Templates: templates,
		Err:       errHandler,
	}

	mux := http.NewServeMux()
	// Статические файлы (CSS, изображения)
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Роуты
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/logout", authHandler.Logout)
	mux.HandleFunc("/create", postHandler.CreatePost)
	mux.HandleFunc("/post/comment", commentHandler.AddComment)
	mux.HandleFunc("/like", likeHandler.Like)
	mux.HandleFunc("/post/", postHandler.GetPost)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			errHandler.NotFound(w, r)
			return
		}
		filterHandler.FilteredPosts(w, r)
	})

	wrappedMux := errHandler.RecoveryMiddleware(mux)
	log.Println("Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", wrappedMux); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
