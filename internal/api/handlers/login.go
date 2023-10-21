package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/maxzhirnov/rewardify/internal/auth"
)

type LoginRequestData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h Handlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	h.logger.Log.Debug("handler HandleLogin starting handle request...")
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if r.Body == nil {
		response := map[string]string{"message": "you have to provide username and password"}
		JSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response := map[string]string{"message": "wrong request format"}
		JSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Парсим тело запроса в структуру
	var requestData LoginRequestData
	if err := json.Unmarshal(body, &requestData); err != nil {
		response := map[string]string{"message": "wrong request format"}
		JSONResponse(w, http.StatusBadRequest, response)
		return
	}

	if requestData.Login == "" || requestData.Password == "" {
		response := map[string]string{"message": "wrong request format"}
		JSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Пытаемся аутентифицировать пользователя
	tokenString, err := h.app.Authenticate(ctx, requestData.Login, requestData.Password)
	if err != nil {
		response := map[string]string{"message": "wrong username or password"}
		JSONResponse(w, http.StatusUnauthorized, response)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    auth.JWTCookeName,
		Value:   tokenString,
		Expires: time.Now().Add(cookieExpirationTime),
	})
	response := map[string]string{"message": "successful login"}
	JSONResponse(w, http.StatusOK, response)
}
