package handlers

import (
	"html/template"
	"net/http"
)

type ErrorHandler struct {
	Templates *template.Template
}

func (h *ErrorHandler) Render(w http.ResponseWriter, status int, msg string) {
	if h == nil || h.Templates == nil {
		http.Error(w, msg, status)
		return
	}
	w.WriteHeader(status)
	err := h.Templates.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Page":   "error",
		"Error":  msg,
		"Status": status,
	})
	if err != nil {
		http.Error(w, "Ошибка при отображении страницы ошибки", http.StatusInternalServerError)
	}
}

func (h *ErrorHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	h.Render(w, http.StatusNotFound, "Страница не найдена")
}

func (h *ErrorHandler) RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				h.Render(w, http.StatusInternalServerError, "Внутренняя ошибка сервера")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
