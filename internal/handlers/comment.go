package handlers

import (
	"database/sql"
	"forum/internal/models"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

type CommentHandler struct {
	DB        *sql.DB
	Templates *template.Template
}

// Добавление комментария
func (h *CommentHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка формы", http.StatusBadRequest)
		return
	}

	userID, _, ok := GetUserFromSession(h.DB, r)
	if !ok {
		SetFlash(w, "flash", "Авторизуйтесь, чтобы комментировать")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	postIDStr := r.FormValue("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID == 0 {
		log.Println("⚠️ Некорректный post_id:", postIDStr)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		SetFlash(w, "flash", "Комментарий не может быть пустым")
		http.Redirect(w, r, "/post/"+postIDStr, http.StatusSeeOther)
		return
	}
	createdAt := time.Now().UTC()

	_, err = h.DB.Exec(`
		INSERT INTO comments (post_id, user_id, content, created_at) 
		VALUES (?, ?, ?, ?)`,
		postID, userID, content, createdAt,
	)
	if err != nil {
		log.Println("Ошибка при добавлении комментария:", err)
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/post/"+postIDStr, http.StatusSeeOther)
}

// Получение комментариев для поста
func GetCommentsByPostID(db *sql.DB, postID int) ([]models.Comment, error) {
	rows, err := db.Query(`
		SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Author, &c.Content, &c.CreatedAt)

		// Подсчёт лайков и дизлайков
		likes, dislikes, err := CountLikes(db, "comment_likes", "comment_id", c.ID)
		if err != nil {
			log.Println("Ошибка подсчёта лайков комментария:", err)
		}
		c.Likes = likes
		c.Dislikes = dislikes

		comments = append(comments, c)
	}
	return comments, nil
}
