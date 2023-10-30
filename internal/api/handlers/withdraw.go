package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	app2 "github.com/maxzhirnov/rewardify/internal/app"
	"github.com/maxzhirnov/rewardify/internal/auth"
)

type WithdrawRequestData struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

func (h Handlers) HandleWithdraw(w http.ResponseWriter, r *http.Request) {
	h.logger.Log.Debug("handler HandleWithdraw starting handle request...")
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	userUUID := ""
	if r.Context().Value(auth.UUIDContextKey) != nil {
		userUUID = r.Context().Value(auth.UUIDContextKey).(string)
	}

	if userUUID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
		return
	}

	if r.Body == nil {
		response := map[string]string{"message": "you should provide withdraw body"}
		JSONResponse(w, http.StatusBadRequest, response)
		return
	}
	defer r.Body.Close()
	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Log.Errorln(err)
		response := map[string]string{"message": "wrong requestData format"}
		JSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Парсим тело запроса в структуру
	var requestData WithdrawRequestData
	if err := json.Unmarshal(body, &requestData); err != nil {
		h.logger.Log.Errorln(err)
		response := map[string]string{"message": "wrong requestData format"}
		JSONResponse(w, http.StatusBadRequest, response)
		return
	}

	err = h.app.CreateWithdrawal(ctx, userUUID, requestData.Order, requestData.Sum)
	if errors.Is(err, app2.ErrInvalidOrderNumber) {
		h.logger.Log.Errorln(err)
		response := map[string]string{"message": "wrong order number"}
		JSONResponse(w, http.StatusUnprocessableEntity, response)
		return
	} else if errors.Is(err, app2.ErrInsufficientFunds) {
		h.logger.Log.Errorln(err)
		response := map[string]string{"message": "insufficient bonus funds"}
		JSONResponse(w, http.StatusPaymentRequired, response)
		return
	} else if err != nil {
		h.logger.Log.Errorln(err)
		response := map[string]string{"message": "something went wrong"}
		JSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	response := map[string]string{"message": "success"}
	JSONResponse(w, http.StatusOK, response)
}
