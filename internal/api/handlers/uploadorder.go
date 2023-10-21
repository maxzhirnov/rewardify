package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	app2 "github.com/maxzhirnov/rewardify/internal/app"
	"github.com/maxzhirnov/rewardify/internal/auth"
)

func (h Handlers) HandleUploadOrder(w http.ResponseWriter, r *http.Request) {
	h.logger.Log.Debug("handler HandleUploadOrder starting handle request...")
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	userUUID := ""
	if r.Context().Value(auth.UUIDContextKey) != nil {
		userUUID = r.Context().Value(auth.UUIDContextKey).(string)
	}

	if userUUID == "" {
		response := map[string]string{"message": "unauthorized"}
		JSONResponse(w, http.StatusUnauthorized, response)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response := map[string]string{"message": "request body couldn't be read"}
		JSONResponse(w, http.StatusBadRequest, response)
		return
	}

	orderNumber := string(body)

	err = h.app.UploadOrder(ctx, orderNumber, userUUID)
	if errors.Is(err, app2.ErrInvalidOrderNumber) {
		response := map[string]string{"message": err.Error()}
		JSONResponse(w, http.StatusUnprocessableEntity, response)
		return
	} else if errors.Is(err, app2.ErrAlreadyCreatedByUser) {
		response := map[string]string{"message": "order already have been create"}
		JSONResponse(w, http.StatusOK, response)
		return
	} else if errors.Is(err, app2.ErrAlreadyCreatedByAnotherUser) {
		response := map[string]string{"message": "this order already have been uploaded by another user"}
		JSONResponse(w, http.StatusConflict, response)
		return
	} else if err != nil {
		response := map[string]string{"message": err.Error()}
		JSONResponse(w, http.StatusInternalServerError, response)
		return
	}
	response := map[string]string{"message": "success"}
	JSONResponse(w, http.StatusAccepted, response)
}
