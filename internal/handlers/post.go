package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"forum/internal/models"
	"log"
)

type PostHandler struct {
	DB        *sql.DB
	Templates *template.Template
}

func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")
	categoryIDs := r.URL.Query()["category"] // ?category=1&category=2

	query := `
                SELECT DISTINCT p.id, p.title, p.content, p.created_at, u.username
                FROM posts p
                JOIN users u ON p.user_id = u.id
                LEFT JOIN post_categories pc ON p.id = pc.post_id
                WHERE 1=1`

	var args []interface{}

	if search != "" {
		query += " AND (p.title LIKE ? OR p.content LIKE ?)"
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern)
	}

	if len(categoryIDs) > 0 {
		query += " AND pc.category_id IN (?" + strings.Repeat(",?", len(categoryIDs)-1) + ")"
		for _, id := range categoryIDs {
			args = append(args, id)
		}
	}

	query += " ORDER BY p.created_at DESC"

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.Author); err == nil {
			// Загрузка категорий для поста
			catRows, err := h.DB.Query(`
				SELECT c.id, c.name
				FROM categories c
				JOIN post_categories pc ON c.id = pc.category_id
				WHERE pc.post_id = ?
			`, post.ID)
			if err == nil {
				for catRows.Next() {
					var cat models.Category
					catRows.Scan(&cat.ID, &cat.Name)
					post.Categories = append(post.Categories, cat)
				}
				catRows.Close()
			}

			// Загрузка лайков/дизлайков (если используешь)
			post.Likes, post.Dislikes, _ = CountLikes(h.DB, "post_likes", "post_id", post.ID)

			posts = append(posts, post)
		}
	}

	// категории для фильтра
	catRows, err := h.DB.Query("SELECT id, name FROM categories")
	var categories []models.Category
	if err == nil {
		defer catRows.Close()
		for catRows.Next() {
			var c models.Category
			catRows.Scan(&c.ID, &c.Name)
			categories = append(categories, c)
		}
	}

	_, username, _ := GetUserFromSession(h.DB, r)

	h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Posts":      posts,
		"Categories": categories,
		"Selected":   categoryIDs,
		"Page":       "index",
		"User":       username,
		"Query":      search,
	})
}

// Получение одного поста по id
func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/post/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var post models.Post
	var author string
	err = h.DB.QueryRow(`
		SELECT p.id, p.user_id, p.title, p.content, p.created_at, u.username
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?
	`, id).Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &author)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	post.Author = author

	likes, dislikes, err := CountLikes(h.DB, "post_likes", "post_id", post.ID)
	if err != nil {
		log.Println("Ошибка подсчёта лайков:", err)
	}
	post.Likes = likes
	post.Dislikes = dislikes

	// Получение категорий поста
	catRows, err := h.DB.Query(`
		SELECT c.id, c.name
		FROM categories c
		JOIN post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = ?
	`, post.ID)
	if err == nil {
		defer catRows.Close()
		for catRows.Next() {
			var cat models.Category
			catRows.Scan(&cat.ID, &cat.Name)
			post.Categories = append(post.Categories, cat)
		}
	}

	comments, _ := GetCommentsByPostID(h.DB, post.ID)
	_, username, _ := GetUserFromSession(h.DB, r)
	flash := GetFlash(w, r, "flash")
	log.Printf(">>> POST #%d: 👍 %d 👎 %d", post.ID, post.Likes, post.Dislikes)
	h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Post":     post,
		"Comments": comments,
		"Author":   author,
		"Page":     "post",
		"Flash":    flash,
		"User":     username,
	})
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userID, username, ok := GetUserFromSession(h.DB, r)
	if !ok {
		SetFlash(w, "flash", "Авторизуйтесь, чтобы создать пост")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		rows, err := h.DB.Query("SELECT id, name FROM categories")
		if err != nil {
			http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var categories []models.Category
		for rows.Next() {
			var cat models.Category
			rows.Scan(&cat.ID, &cat.Name)
			categories = append(categories, cat)
		}

		h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
			"Page":       "create",
			"Categories": categories,
			"User":       username,
		})
		return
	}

	// POST-запрос
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка формы", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	catIDs := r.Form["categories"]

	if title == "" || content == "" || len(catIDs) == 0 {
		h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
			"Page":  "create",
			"Error": "Заполните все поля",
		})
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	res, err := tx.Exec("INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)", userID, title, content)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Ошибка создания поста", http.StatusInternalServerError)
		return
	}
	postID, _ := res.LastInsertId()

	for _, catID := range catIDs {
		tx.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, catID)
	}
	tx.Commit()

	http.Redirect(w, r, fmt.Sprintf("/post/%d", postID), http.StatusSeeOther)
}
