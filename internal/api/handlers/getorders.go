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
	OrderNumber string `json:"number"`
	Status      string `json:"status"`
	Accrual     int    `json:"accrual"`
	UploadedAt  string `json:"uploaded_at"`
}

func (h Handlers) HandleGetOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")

	userUUID := r.Context().Value("uuid").(string)
	if userUUID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
		return
	}

	orders, err := h.app.GetAllOrders(ctx, userUUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unauthorized"))
		return
	}

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
			Accrual:     int(o.BonusesAccrued),
			UploadedAt:  o.CreatedAt.Format(time.RFC3339),
		}
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrong"))
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}
