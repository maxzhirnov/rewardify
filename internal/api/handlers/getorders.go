package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type GetOrdersResponseData struct {
	Orders []OrderDTO `json:"orders"`
}

type OrderDTO struct {
	OrderNumber string  `json:"number"`
	Status      string  `json:"status"`
	Accrual     float32 `json:"accrual"`
	UploadedAt  string  `json:"uploaded_at"`
}

func (h Handlers) HandleGetOrders(w http.ResponseWriter, r *http.Request) {
	h.logger.Log.Debug("handler HandleGetOrders starting handle request...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")

	userUUID := r.Context().Value("uuid").(string)
	if userUUID == "" {

		response := map[string]string{"message": "unauthorized"}
		JSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	orders, err := h.app.GetAllOrders(ctx, userUUID)
	if err != nil {
		response := map[string]string{"message": "something went wrong"}
		JSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	h.logger.Log.Debugln("GetAllOrders", orders)

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("no content"))
		return
	}

	response := make([]OrderDTO, len(orders))

	for i, o := range orders {
		response[i] = OrderDTO{
			OrderNumber: o.OrderNumber,
			Status:      string(o.BonusAccrualStatus),
			Accrual:     o.BonusesAccrued,
			UploadedAt:  o.CreatedAt.Format(time.RFC3339),
		}
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		response := map[string]string{"message": "something went wrong"}
		JSONResponse(w, http.StatusInternalServerError, response)
		return
	}
	h.logger.Log.Debugln("getAllOrders json", string(responseJSON))

	JSONResponse(w, http.StatusOK, responseJSON)
}
