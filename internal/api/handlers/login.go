package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type LoginRequestData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponseData struct {
	Message string `json:"message"`
}

func (h Handlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Создаем инстанс респонса и устанавливаем хэдер Content-Type
	response := new(LoginResponseData)
	w.Header().Set("Content-Type", "application/json")

	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = "Wrong requestData format"
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			w.Write([]byte(response.Message))
			return
		}
		w.Write(jsonResponse)
		return
	}

	// Парсим тело запроса в структуру
	var requestData LoginRequestData
	if err := json.Unmarshal(body, &requestData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = "wrong requestData format"
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			w.Write([]byte(response.Message))
			return
		}
		w.Write(jsonResponse)
		return
	}

	// Пытаемся аутентифицировать пользователя
	tokenString, err := h.app.Authenticate(ctx, requestData.Login, requestData.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		response.Message = "wrong username or password"
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			w.Write([]byte(response.Message))
			return
		}
		w.Write(jsonResponse)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: time.Now().Add(cookieExpirationTime),
	})

	w.WriteHeader(http.StatusOK)
	response.Message = "successful login"
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		w.Write([]byte(response.Message))
		return
	}
	w.Write(jsonResponse)
	return
}
