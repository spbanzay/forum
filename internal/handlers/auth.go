package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"log"

	"html/template"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB        *sql.DB
	Templates *template.Template
}

// Регистрация пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		_, username, _ := GetUserFromSession(h.DB, r)
		flash := GetFlash(w, r, "flash")
		h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
			"Error": nil,
			"Page":  "register",
			"Flash": flash,
			"User":  username,
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderError(w, "Ошибка формы")
		return
	}

	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")

	if email == "" || username == "" || password == "" {
		h.renderError(w, "Все поля обязательны")
		return
	}

	// Проверка на уникальность email
	var exists int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&exists)
	if err != nil {
		h.renderError(w, "Ошибка базы данных")
		return
	}
	if exists > 0 {
		h.renderError(w, "Email уже занят")
		return
	}

	// Проверка на уникальность username
	err = h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&exists)
	if err != nil {
		h.renderError(w, "Ошибка базы данных")
		return
	}
	if exists > 0 {
		h.renderError(w, "Имя пользователя уже занято")
		return
	}

	// Хеширование пароля
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		h.renderError(w, "Ошибка шифрования пароля")
		return
	}

	_, err = h.DB.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", email, username, string(hashed))
	if err != nil {
		h.renderError(w, "Ошибка создания пользователя")
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Вход пользователя
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		_, username, _ := GetUserFromSession(h.DB, r)
		flash := GetFlash(w, r, "flash")
		h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
			"Error": nil,
			"Page":  "login",
			"Flash": flash,
			"User":  username,
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderError(w, "Ошибка формы")
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	var id int
	var hashed string
	err := h.DB.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&id, &hashed)
	if err == sql.ErrNoRows {
		h.renderError(w, "Пользователь не найден")
		return
	} else if err != nil {
		h.renderError(w, "Ошибка базы данных")
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) != nil {
		h.renderError(w, "Неверный пароль")
		return
	}

	// Создание сессии
	sessionID := uuid.New().String()
	expires := time.Now().Add(24 * time.Hour)
	_, err = h.DB.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, id, expires)
	if err != nil {
		h.renderError(w, "Ошибка создания сессии")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  expires,
		HttpOnly: true,
		Path:     "/",
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Выход пользователя
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		h.DB.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
		cookie.Expires = time.Unix(0, 0)
		http.SetCookie(w, cookie)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Вспомогательная функция для вывода ошибок
func (h *AuthHandler) renderError(w http.ResponseWriter, msg string) {
	h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Error": msg,
	})
}

func GetUserFromSession(db *sql.DB, r *http.Request) (int, string, bool) {
	// Пытаемся извлечь cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Println("COOKIE НЕ НАЙДЕНА")
		return 0, "", false
	}
	sessionID := cookie.Value
	log.Println("COOKIE session_id =", sessionID)

	// Проверяем сессию и получаем user_id
	var userID int
	var expiresAt time.Time
	err = db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE id = ?", sessionID).
		Scan(&userID, &expiresAt)
	if err != nil {
		log.Println("СЕССИЯ НЕ НАЙДЕНА В БД или ОШИБКА:", err)
		return 0, "", false
	}
	log.Printf("Найдена сессия: user_id=%d, expiresAt=%v, now=%v", userID, expiresAt, time.Now())

	if time.Now().UTC().After(expiresAt.UTC()) {
		log.Println("СЕССИЯ ИСТЕКЛА")
		return 0, "", false
	}

	// Получаем username по user_id
	var username string
	err = db.QueryRow("SELECT username FROM users WHERE id = ?", userID).
		Scan(&username)
	if err != nil {
		log.Println("USERNAME НЕ НАЙДЕН:", err)
		return 0, "", false
	}
	log.Printf("USER %s ПОДТВЕРЖДЁН", username)

	return userID, username, true
}

func SetFlash(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:  name,
		Value: value,
		Path:  "/",
	})
}

func GetFlash(w http.ResponseWriter, r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:   name,
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		return cookie.Value
	}
	return ""
}
