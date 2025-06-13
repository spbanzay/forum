package main

import (
	"database/sql"
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
	})

	templates, err = templates.ParseGlob(filepath.Join("templates", "*.html"))
	for _, tmpl := range templates.Templates() {
		log.Println("Загружен шаблон:", tmpl.Name())
	}

	if err != nil {
		log.Fatal("Ошибка парсинга шаблонов:", err)
	}

	commentHandler := handlers.CommentHandler{
		DB:        db,
		Templates: templates,
	}

	likeHandler := handlers.LikeHandler{
		DB: db,
	}

	filterHandler := handlers.FilterHandler{
		DB:        db,
		Templates: templates,
	}

	postHandler := handlers.PostHandler{
		DB:        db,
		Templates: templates,
	}

	authHandler := handlers.AuthHandler{
		DB:        db,
		Templates: templates,
	}

	// Статические файлы (CSS, изображения)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Роуты
	http.HandleFunc("/", postHandler.ListPosts)
	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/login", authHandler.Login)
	http.HandleFunc("/logout", authHandler.Logout)
	http.HandleFunc("/create", postHandler.CreatePost)
	http.HandleFunc("/post/", postHandler.GetPost)
	http.HandleFunc("/post/comment", commentHandler.AddComment)
	http.HandleFunc("/like", likeHandler.Like)
	http.HandleFunc("/filter/category", filterHandler.ByCategory)
	http.HandleFunc("/filter/liked", filterHandler.ByLiked)

	log.Println("Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
