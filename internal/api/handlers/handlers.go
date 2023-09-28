package handlers

import (
	"net/http"

	"github.com/maxzhirnov/rewardify/internal/logger"
)

type Handlers struct {
	logger *logger.Logger
}

func NewHandlers(l *logger.Logger) *Handlers {
	return &Handlers{
		logger: l,
	}
}

func (h Handlers) HandlePing(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("pong"))
	if err != nil {
		return
	}
}

func (h Handlers) HandleRegister(writer http.ResponseWriter, request *http.Request) {

}

func (h Handlers) HandleLogin(writer http.ResponseWriter, request *http.Request) {

}

func (h Handlers) HandleOrdersUpload(writer http.ResponseWriter, request *http.Request) {

}

func (h Handlers) HandleOrdersList(writer http.ResponseWriter, request *http.Request) {

}

func (h Handlers) HandleGetBalance(writer http.ResponseWriter, request *http.Request) {

}

func (h Handlers) HandleWithdraw(writer http.ResponseWriter, request *http.Request) {

}

func (h Handlers) HandleListAllWithdraws(writer http.ResponseWriter, request *http.Request) {

}
