package handlers

import (
	"context"
	"net/http"
	"time"
)

func (h Handlers) HandleWithdraw(w http.ResponseWriter, r *http.Request) {
	h.logger.Log.Debug("handler HandleWithdraw starting handle request...")
	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
}
