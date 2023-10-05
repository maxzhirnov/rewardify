package accrual

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/maxzhirnov/rewardify/internal/logger"
)

var (
	errBadRequest      = errors.New("bad request")
	errTooManyRequests = errors.New("too many requests")
	errNotRegistered   = errors.New("order hasn't been registered in system")
)

type APIResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type APIWrapper struct {
	addr   string
	client *http.Client
	logger *logger.Logger
}

func NewAPIWrapper(addr string, client *http.Client, logger *logger.Logger) *APIWrapper {
	return &APIWrapper{
		addr:   addr,
		client: client,
		logger: logger,
	}
}

func (a APIWrapper) Fetch(orderNumber string) (*APIResponse, error) {
	apiURL, err := url.JoinPath(a.addr, "/api/orders/", orderNumber)
	if err != nil {
		errWrapped := fmt.Errorf("error building api apiURL: %w", err)
		a.logger.Log.Error(errWrapped)
		return nil, errWrapped
	}

	resp, err := a.client.Get(apiURL)
	if err != nil {
		errWrapped := fmt.Errorf("error fetching api: %w", err)
		a.logger.Log.Error(errWrapped)
		return nil, errWrapped
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, errTooManyRequests
	}

	if resp.StatusCode == 204 {
		return nil, errNotRegistered
	}

	if resp.StatusCode != 200 {
		return nil, errBadRequest
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	response := &APIResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}