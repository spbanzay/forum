package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Тип сущности для лайка: "post" или "comment"
type LikeHandler struct {
	DB *sql.DB
}

// Обработчик для лайка/дизлайка поста или комментария
func (h *LikeHandler) Like(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	userID, _, ok := GetUserFromSession(h.DB, r)
	if !ok {
		http.Error(w, "Только для авторизованных пользователей", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка формы", http.StatusBadRequest)
		return
	}

	typ := r.FormValue("type")      // "post" или "comment"
	idStr := r.FormValue("id")      // post_id или comment_id
	action := r.FormValue("action") // "like" или "dislike"
	targetID, err := strconv.Atoi(idStr)
	if err != nil || (action != "like" && action != "dislike") {
		http.Error(w, "Некорректные параметры", http.StatusBadRequest)
		return
	}

	isLike := action == "like"

	var (
		table     string
		column    string
		deleteSQL string
		insertSQL string
	)

	switch typ {
	case "post":
		table = "post_likes"
		column = "post_id"
	case "comment":
		table = "comment_likes"
		column = "comment_id"
	default:
		http.Error(w, "Неверный тип", http.StatusBadRequest)
		return
	}

	// Генерация SQL под конкретную таблицу
	deleteSQL = fmt.Sprintf("DELETE FROM %s WHERE %s = ? AND user_id = ?", table, column)
	insertSQL = fmt.Sprintf("INSERT INTO %s (%s, user_id, is_like) VALUES (?, ?, ?)", table, column)

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "Ошибка транзакции", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(deleteSQL, targetID, userID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Ошибка при удалении лайка", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(insertSQL, targetID, userID, isLike)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Ошибка при добавлении лайка", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Ошибка при коммите", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	log.Printf("[LIKE] redirect to: %s", r.Header.Get("Referer"))
}

// Получить количество лайков и дизлайков для сущности
func CountLikes(db *sql.DB, table string, column string, targetID int) (likes, dislikes int, err error) {
	likeQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ? AND is_like = 1", table, column)
	dislikeQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ? AND is_like = 0", table, column)

	err = db.QueryRow(likeQuery, targetID).Scan(&likes)
	if err != nil {
		return
	}
	err = db.QueryRow(dislikeQuery, targetID).Scan(&dislikes)
	return
}

// Получить postID по commentID (реализуйте в зависимости от вашей схемы)
func GetPostIDByCommentID(db *sql.DB, commentID int) int {
	var postID int
	_ = db.QueryRow("SELECT post_id FROM comments WHERE id = ?", commentID).Scan(&postID)
	return postID
}
