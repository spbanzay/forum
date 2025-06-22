package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"log"

	"html/template"

	"regexp"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB        *sql.DB
	Templates *template.Template
	Err       *ErrorHandler
}

// Проверка формата email
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// Проверка формата username
func isValidUsername(username string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
	return re.MatchString(username)
}

// Минимальная длина пароля — 6 символов
func isValidPassword(password string) bool {
	return len(password) >= 6
}

// Регистрация пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		_, username, _ := GetUserFromSession(h.DB, r)
		flash := GetFlash(w, r, "flash")
		h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
			"Page":       "register",
			"User":       username,
			"Flash":      flash,
			"FormErrors": map[string]string{},
			"FormValues": map[string]string{},
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		h.Err.Render(w, http.StatusBadRequest, "Ошибка формы")
		return
	}

	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")

	formErrors := make(map[string]string)

	if email == "" {
		formErrors["Email"] = "Введите email"
	} else if !isValidEmail(email) {
		formErrors["Email"] = "Некорректный формат email"
	}

	if username == "" {
		formErrors["Username"] = "Введите имя пользователя"
	} else if !isValidUsername(username) {
		formErrors["Username"] = "Имя может содержать только буквы, цифры и подчёркивания (3-20 символов)"
	}

	if password == "" {
		formErrors["Password"] = "Введите пароль"
	} else if !isValidPassword(password) {
		formErrors["Password"] = "Пароль должен быть не менее 6 символов"
	}

	// Проверка email в базе
	var exists int
	if email != "" {
		err := h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&exists)
		if err != nil {
			h.Err.Render(w, http.StatusInternalServerError, "Ошибка базы данных")
			return
		}
		if exists > 0 {
			formErrors["Email"] = "Email уже занят"
		}
	}

	// Проверка username в базе
	if username != "" {
		err := h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&exists)
		if err != nil {
			h.Err.Render(w, http.StatusInternalServerError, "Ошибка базы данных")
			return
		}
		if exists > 0 {
			formErrors["Username"] = "Имя пользователя уже занято"
		}
	}

	if len(formErrors) > 0 {
		h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
			"Page":       "register",
			"FormErrors": formErrors,
			"FormValues": map[string]string{
				"Email":    email,
				"Username": username,
			},
		})
		return
	}

	// Хеширование пароля
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		h.Err.Render(w, http.StatusInternalServerError, "Ошибка шифрования пароля")
		return
	}

	res, err := h.DB.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", email, username, string(hashed))
	if err != nil {
		h.Err.Render(w, http.StatusInternalServerError, "Ошибка создания пользователя")
		return
	}

	userID, err := res.LastInsertId()
	if err != nil {
		h.Err.Render(w, http.StatusInternalServerError, "Ошибка создания пользователя")
		return
	}

	// Создание сессии
	sessionID := uuid.New().String()
	expires := time.Now().Add(24 * time.Hour)
	_, err = h.DB.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, userID, expires)
	if err != nil {
		h.Err.Render(w, http.StatusInternalServerError, "Ошибка создания сессии")
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

// Вход пользователя
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		_, username, _ := GetUserFromSession(h.DB, r)
		flash := GetFlash(w, r, "flash")
		h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
			"Page":       "login",
			"User":       username,
			"Flash":      flash,
			"FormErrors": map[string]string{},
			"FormValues": map[string]string{},
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		h.Err.Render(w, http.StatusBadRequest, "Ошибка формы")
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	formErrors := make(map[string]string)

	if email == "" {
		formErrors["Email"] = "Введите email"
	} else if !isValidEmail(email) {
		formErrors["Email"] = "Некорректный формат email"
	}

	if password == "" {
		formErrors["Password"] = "Введите пароль"
	}

	if len(formErrors) > 0 {
		h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
			"Page":       "login",
			"FormErrors": formErrors,
			"FormValues": map[string]string{
				"Email": email,
			},
		})
		return
	}

	// Поиск пользователя
	var id int
	var hashed string
	err := h.DB.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&id, &hashed)
	if err == sql.ErrNoRows {
		formErrors["Email"] = "Пользователь не найден"
		h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
			"Page":       "login",
			"FormErrors": formErrors,
			"FormValues": map[string]string{"Email": email},
		})
		return
	} else if err != nil {
		h.Err.Render(w, http.StatusInternalServerError, "Ошибка базы данных")
		return
	}

	// Проверка пароля
	if bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) != nil {
		formErrors["Password"] = "Неверный пароль"
		h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
			"Page":       "login",
			"FormErrors": formErrors,
			"FormValues": map[string]string{"Email": email},
		})
		return
	}

	// Удаление старых сессий
	_, err = h.DB.Exec("DELETE FROM sessions WHERE user_id = ?", id)
	if err != nil {
		h.Err.Render(w, http.StatusInternalServerError, "Ошибка удаления старых сессий")
		return
	}

	// Создание новой сессии
	sessionID := uuid.New().String()
	expires := time.Now().Add(24 * time.Hour)
	_, err = h.DB.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, id, expires)
	if err != nil {
		h.Err.Render(w, http.StatusInternalServerError, "Ошибка создания сессии")
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
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
