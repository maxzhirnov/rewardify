package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type GetBalanceResponseData struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func (h Handlers) HandleGetBalance(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")

	userUUID := r.Context().Value("uuid").(string)
	if userUUID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
		return
	}

	balance, err := h.app.GetBalance(ctx, userUUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrong"))
		return
	}

	response := GetBalanceResponseData{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrong"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
