package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
)

// Тип сущности для лайка: "post" или "comment"
type LikeHandler struct {
	DB  *sql.DB
	Err *ErrorHandler
}

// Обработчик для лайка/дизлайка поста или комментария
func (h *LikeHandler) Like(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	userID, _, ok := GetUserFromSession(h.DB, r)
	if !ok {
		h.Err.Render(w, http.StatusUnauthorized, "Только для авторизованных пользователей")
		return
	}

	if err := r.ParseForm(); err != nil {
		h.Err.Render(w, http.StatusBadRequest, "Ошибка формы")
		return
	}

	typ := r.FormValue("type")      // "post" или "comment"
	idStr := r.FormValue("id")      // post_id или comment_id
	action := r.FormValue("action") // "like" или "dislike"
	targetID, err := strconv.Atoi(idStr)
	if err != nil || (action != "like" && action != "dislike") {
		h.Err.Render(w, http.StatusBadRequest, "Некорректные параметры")
		return
	}

	isLike := action == "like"

	var table, column string

	switch typ {
	case "post":
		table = "post_likes"
		column = "post_id"
	case "comment":
		table = "comment_likes"
		column = "comment_id"
	default:
		h.Err.Render(w, http.StatusBadRequest, "Неверный тип")
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		h.Err.Render(w, http.StatusInternalServerError, "Ошибка транзакции")
		return
	}

	// Проверим, был ли лайк/дизлайк ранее
	var currentValue bool
	err = tx.QueryRow(
		fmt.Sprintf("SELECT is_like FROM %s WHERE %s = ? AND user_id = ?", table, column),
		targetID, userID,
	).Scan(&currentValue)

	if err == sql.ErrNoRows {
		// Ещё не было лайка — добавим
		_, err = tx.Exec(
			fmt.Sprintf("INSERT INTO %s (%s, user_id, is_like) VALUES (?, ?, ?)", table, column),
			targetID, userID, isLike,
		)
		if err != nil {
			tx.Rollback()
			h.Err.Render(w, http.StatusInternalServerError, "Ошибка при добавлении лайка")
			return
		}
	} else if err == nil {
		if currentValue == isLike {
			// Повторное нажатие — удалим
			_, err = tx.Exec(
				fmt.Sprintf("DELETE FROM %s WHERE %s = ? AND user_id = ?", table, column),
				targetID, userID,
			)
			if err != nil {
				tx.Rollback()
				h.Err.Render(w, http.StatusInternalServerError, "Ошибка при удалении лайка")
				return
			}
		} else {
			// Меняем статус
			_, err = tx.Exec(
				fmt.Sprintf("UPDATE %s SET is_like = ? WHERE %s = ? AND user_id = ?", table, column),
				isLike, targetID, userID,
			)
			if err != nil {
				tx.Rollback()
				h.Err.Render(w, http.StatusInternalServerError, "Ошибка при обновлении лайка")
				return
			}
		}
	} else {
		tx.Rollback()
		h.Err.Render(w, http.StatusInternalServerError, "Ошибка проверки состояния")
		return
	}

	if err := tx.Commit(); err != nil {
		h.Err.Render(w, http.StatusInternalServerError, "Ошибка при коммите")
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
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
