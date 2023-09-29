package handlers

import (
	"encoding/json"
	"net/http"
)

type PingResponseData struct {
	Username string `json:"username"`
	UUID     string `json:"uuid"`
}

func (h Handlers) HandlePing(w http.ResponseWriter, r *http.Request) {
	response := PingResponseData{}
	w.Header().Set("Content-Type", "application/json")

	username := r.Context().Value("username")
	uuid := r.Context().Value("uuid")

	if username != nil && uuid != nil {
		response.Username = username.(string)
		response.UUID = uuid.(string)
	}

	if err := h.app.Ping(); err != nil {
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
	return
}
