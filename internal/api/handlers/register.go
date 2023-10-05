package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	a "github.com/maxzhirnov/rewardify/internal/app"
)

const (
	tokenCookieName = "token"
)

type RegisterRequestData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h Handlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if r.Body == nil {
		response := map[string]string{"message": "you should provide request body"}
		JSONResponse(w, http.StatusBadRequest, response)
		return
	}
	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response := map[string]string{"message": "wrong requestData format"}
		JSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Парсим тело запроса в структуру
	var requestData RegisterRequestData
	if err := json.Unmarshal(body, &requestData); err != nil {
		response := map[string]string{"message": "wrong requestData format"}
		JSONResponse(w, http.StatusBadRequest, response)
		return
	}

	if requestData.Login == "" || requestData.Password == "" {
		response := map[string]string{"message": "user or password shouldn't be empty strings"}
		JSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Пытаемся зарегистрировать пользователя
	if err := h.app.Register(ctx, requestData.Login, requestData.Password); err != nil {
		if errors.Is(err, a.ErrUserAlreadyExist) {
			response := map[string]string{"message": "username already exists"}
			JSONResponse(w, http.StatusConflict, response)
			return
		}
		response := map[string]string{"message": "something went wrong"}
		JSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	// В случае успешной регистрации сразу аутентифицируем пользователя
	tokenString, err := h.app.Authenticate(ctx, requestData.Login, requestData.Password)
	if err != nil {
		response := map[string]string{"message": "something went wrong"}
		JSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    tokenCookieName,
		Value:   tokenString,
		Expires: time.Now().Add(cookieExpirationTime),
	})
	response := map[string]string{"message": "user was successfully registered"}
	JSONResponse(w, http.StatusOK, response)
	return
}
