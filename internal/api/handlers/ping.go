package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/maxzhirnov/rewardify/internal/auth"
)

type PingResponseData struct {
	Username string `json:"username"`
	UUID     string `json:"uuid"`
}

func (h Handlers) HandlePing(w http.ResponseWriter, r *http.Request) {
	h.logger.Log.Debug("handler HandlePing starting handle request...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	response := PingResponseData{}
	w.Header().Set("Content-Type", "application/json")

	username := r.Context().Value(auth.UsernameContextKey)
	uuid := r.Context().Value(auth.UUIDContextKey)

	if username != nil && uuid != nil {
		response.Username = username.(string)
		response.UUID = uuid.(string)
	}

	if err := h.app.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			h.logger.Log.Error(err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		w.Write([]byte("OK"))
		return
	}
	w.Write(jsonResponse)
}
