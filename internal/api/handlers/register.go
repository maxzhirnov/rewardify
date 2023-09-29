package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	a "github.com/maxzhirnov/rewardify/internal/app"
)

type RegisterRequestData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterResponseData struct {
	Message string `json:"message"`
}

func (h Handlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	// Создаем инстанс респонса и устанавливаем хэдер Content-Type
	response := new(RegisterResponseData)
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
	var requestData RegisterRequestData
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

	// Пытаемся зарегистрировать пользователя
	if err := h.app.Register(requestData.Login, requestData.Password); err != nil {
		if errors.Is(err, a.ErrUserAlreadyExist) {
			w.WriteHeader(http.StatusConflict)
			response.Message = "username already exists"
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				w.Write([]byte(response.Message))
				return
			}
			w.Write(jsonResponse)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		response.Message = "something went wrong"
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			w.Write([]byte(response.Message))
			return
		}
		w.Write(jsonResponse)
		return
	}

	// В случае успешной регистрации сразу аутентифицируем пользователя
	tokenString, err := h.app.Authenticate(requestData.Login, requestData.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.Message = "something went wrong"
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
	response.Message = "user was successfully registered"
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		w.Write([]byte(response.Message))
		return
	}
	w.Write(jsonResponse)
	return
}
