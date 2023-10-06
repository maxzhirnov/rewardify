package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	app2 "github.com/maxzhirnov/rewardify/internal/app"
)

type UploadOrderRequestData string

type UploadOrderResponseData struct {
	Message string `json:"message"`
}

func (h Handlers) HandleUploadOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	response := new(UploadOrderResponseData)
	userUUID := r.Context().Value("uuid").(string)

	if userUUID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = "request body couldn't be read"
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			w.Write([]byte(response.Message))
			return
		}
		w.Write(jsonResponse)
		return
	}

	orderNumber := string(body)

	err = h.app.UploadOrder(ctx, orderNumber, userUUID)
	if errors.Is(err, app2.ErrInvalidOrderNumber) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		response.Message = err.Error()
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			w.Write([]byte(response.Message))
			return
		}
		w.Write(jsonResponse)
		return
	} else if errors.Is(err, app2.ErrAlreadyCreatedByUser) {
		w.WriteHeader(http.StatusOK)
		response.Message = "order already have been created"
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			w.Write([]byte(response.Message))
			return
		}
		w.Write(jsonResponse)
		return
	} else if errors.Is(err, app2.ErrAlreadyCreatedByAnotherUser) {
		w.WriteHeader(http.StatusConflict)
		response.Message = "this order already have been uploaded by another user"
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			w.Write([]byte(response.Message))
			return
		}
		w.Write(jsonResponse)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.Message = err.Error()
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			w.Write([]byte(response.Message))
			return
		}
		w.Write(jsonResponse)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	response.Message = "success"
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		w.Write([]byte(response.Message))
		return
	}
	w.Write(jsonResponse)
}
