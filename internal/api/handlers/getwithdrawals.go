package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/maxzhirnov/rewardify/internal/api/middlewares"
)

type WithdrawalDTO struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func (h Handlers) HandleListAllWithdrawals(w http.ResponseWriter, r *http.Request) {
	h.logger.Log.Debug("handler HandleListAllWithdrawals starting handle request...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	w.Header().Set("Content-Type", "application/json")
	userUUID := r.Context().Value(middlewares.UUIDContextKey).(string)
	if userUUID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
		return
	}

	withdrawals, err := h.app.GetAllWithdrawals(ctx, userUUID)
	if err != nil {
		response := map[string]string{"message": "something went wrong"}
		JSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	if len(withdrawals) == 0 {
		response := map[string]string{"message": "no withdrawals"}
		JSONResponse(w, http.StatusNoContent, response)
		return
	}

	response := make([]WithdrawalDTO, len(withdrawals))

	for i, w := range withdrawals {
		response[i].Order = w.OrderNumber
		response[i].Sum = w.Amount
		response[i].ProcessedAt = w.CreatedAt.Format(time.RFC3339)
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		response := map[string]string{"message": "something went wrong"}
		JSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}
