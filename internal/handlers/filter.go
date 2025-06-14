package handlers

import (
	"database/sql"
	"forum/internal/models"
	"html/template"
	"net/http"
	"strings"
)

type FilterHandler struct {
	DB        *sql.DB
	Templates *template.Template
}

func (h *FilterHandler) FilteredPosts(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	query := r.FormValue("q")
	selectedCategories := r.Form["category"]
	liked := r.FormValue("liked") == "1"

	userID, username, _ := GetUserFromSession(h.DB, r)

	posts, err := GetFilteredPosts(h.DB, query, selectedCategories, liked, userID)
	if err != nil {
		http.Error(w, "Ошибка загрузки постов", http.StatusInternalServerError)
		return
	}

	categories := LoadAllCategories(h.DB)

	h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Page":       "index",
		"Posts":      posts,
		"Categories": categories,
		"Selected":   selectedCategories,
		"User":       username,
		"Query":      query,
		"LikedView":  liked,
	})
}

func GetFilteredPosts(db *sql.DB, query string, selectedCategories []string, liked bool, userID int) ([]models.Post, error) {
	var posts []models.Post
	var args []interface{}
	conditions := []string{}

	// Поиск по тексту
	if query != "" {
		conditions = append(conditions, "(p.title LIKE ? OR p.content LIKE ?)")
		likePattern := "%" + query + "%"
		args = append(args, likePattern, likePattern)
	}

	// Фильтрация по категориям
	if len(selectedCategories) > 0 {
		placeholders := strings.Repeat("?,", len(selectedCategories))
		placeholders = placeholders[:len(placeholders)-1] // удалить последнюю запятую
		conditions = append(conditions, `
			EXISTS (
				SELECT 1 FROM post_categories pc
				WHERE pc.post_id = p.id AND pc.category_id IN (`+placeholders+`)
			)
		`)
		for _, catID := range selectedCategories {
			args = append(args, catID)
		}
	}

	// Фильтрация по лайкам
	if liked {
		conditions = append(conditions, `
			EXISTS (
				SELECT 1 FROM post_likes pl
				WHERE pl.post_id = p.id AND pl.user_id = ? AND pl.is_like = TRUE
			)
		`)
		args = append(args, userID)
	}

	queryStr := `
		SELECT DISTINCT p.id, p.user_id, p.title, p.content, p.created_at, u.username
		FROM posts p
		JOIN users u ON p.user_id = u.id
	`
	if len(conditions) > 0 {
		queryStr += " WHERE " + strings.Join(conditions, " AND ")
	}
	queryStr += " ORDER BY p.created_at DESC"

	rows, err := db.Query(queryStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &post.Author); err != nil {
			continue
		}

		// Подсчёт лайков
		post.Likes, post.Dislikes, _ = CountLikes(db, "post_likes", "post_id", post.ID)

		// Категории поста
		cats, err := loadCategoriesForPost(db, post.ID)
		if err == nil {
			post.Categories = cats
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func loadCategoriesForPost(db *sql.DB, postID int) ([]models.Category, error) {
	rows, err := db.Query(`
		SELECT c.id, c.name
		FROM categories c
		JOIN post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = ?
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err == nil {
			cats = append(cats, c)
		}
	}
	return cats, nil
}

func LoadAllCategories(db *sql.DB) []models.Category {
	rows, err := db.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err == nil {
			categories = append(categories, c)
		}
	}
	return categories
}
