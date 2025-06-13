package handlers

import (
	"database/sql"
	"forum/internal/models"
	"html/template"
	"net/http"
)

type FilterHandler struct {
	DB        *sql.DB
	Templates *template.Template
}

// Фильтрация по категориям
func (h *FilterHandler) ByCategory(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ids := r.Form["id"] // массив выбранных ID категорий

	if len(ids) == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Подготовка SQL-запроса
	placeholders := ""
	args := []interface{}{}
	for i, idStr := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, idStr)
	}

	rows, err := h.DB.Query(`
		SELECT DISTINCT p.id, p.user_id, p.title, p.content, p.created_at, u.username
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN post_categories pc ON p.id = pc.post_id
		WHERE pc.category_id IN (`+placeholders+`)
		ORDER BY p.created_at DESC
	`, args...)
	if err != nil {
		http.Error(w, "Ошибка запроса к базе", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &post.Author)
		if err != nil {
			continue
		}

		// Лайки
		post.Likes, post.Dislikes, _ = CountLikes(h.DB, "post_likes", "post_id", post.ID)

		// Категории
		catRows, _ := h.DB.Query(`
			SELECT c.id, c.name
			FROM categories c
			JOIN post_categories pc ON c.id = pc.category_id
			WHERE pc.post_id = ?
		`, post.ID)
		for catRows.Next() {
			var c models.Category
			catRows.Scan(&c.ID, &c.Name)
			post.Categories = append(post.Categories, c)
		}
		catRows.Close()

		posts = append(posts, post)
	}

	// Загрузка всех категорий
	allCategories := []models.Category{}
	catRows, _ := h.DB.Query("SELECT id, name FROM categories")
	for catRows.Next() {
		var c models.Category
		catRows.Scan(&c.ID, &c.Name)
		allCategories = append(allCategories, c)
	}
	catRows.Close()

	// Преобразование выбранных id в []string для шаблона
	selected := append([]string{}, ids...)

	_, username, _ := GetUserFromSession(h.DB, r)
	h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Page":       "index",
		"Posts":      posts,
		"Categories": allCategories,
		"Selected":   selected,
		"User":       username,
	})
}

// Фильтрация по понравившимся постам пользователя
func (h *FilterHandler) ByLiked(w http.ResponseWriter, r *http.Request) {
	userID, username, ok := GetUserFromSession(h.DB, r)
	if !ok {
		SetFlash(w, "flash", "Авторизуйтесь, чтобы фильтровать по лайкам")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Получаем посты, которым пользователь поставил лайк
	rows, err := h.DB.Query(`
		SELECT p.id, p.title, p.content, p.created_at, u.username
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN post_likes pl ON pl.post_id = p.id
		WHERE pl.user_id = ? AND pl.is_like = TRUE
		ORDER BY p.created_at DESC
	`, userID)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.Author); err == nil {
			// Подсчёт лайков и категорий (если нужно)
			post.Likes, post.Dislikes, _ = CountLikes(h.DB, "post_likes", "post_id", post.ID)

			// Категории
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

			posts = append(posts, post)
		}
	}

	// Все категории (для фильтра)
	categories := []models.Category{}
	catRows, err := h.DB.Query("SELECT id, name FROM categories")
	if err == nil {
		defer catRows.Close()
		for catRows.Next() {
			var c models.Category
			catRows.Scan(&c.ID, &c.Name)
			categories = append(categories, c)
		}
	}

	h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Page":       "index",
		"Posts":      posts,
		"Categories": categories,
		"User":       username,
		"LikedView":  true,
	})
}
