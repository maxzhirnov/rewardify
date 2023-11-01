package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/maxzhirnov/rewardify/internal/logger"
)

var (
	errBadRequest      = errors.New("bad request")
	errTooManyRequests = errors.New("too many requests")
	errNotRegistered   = errors.New("order hasn't been registered in system")
)

const retryAfterDefault = 30

type APIResponse struct {
	Order      string  `json:"order"`
	Status     string  `json:"status"`
	Accrual    float32 `json:"accrual"`
	RetryAfter int
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

func (a APIWrapper) Fetch(ctx context.Context, orderNumber string) (*APIResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	apiURL, err := url.JoinPath(a.addr, "api", "orders", orderNumber)
	if err != nil {
		a.logger.Log.Error(err)
		return nil, err
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		a.logger.Log.Error(err)
		return nil, err
	}

	req = req.WithContext(ctx)

	resp, err := a.client.Do(req)
	if err != nil {
		errWrapped := fmt.Errorf("error fetching api: %w", err)
		a.logger.Log.Error(errWrapped)
		return nil, errWrapped
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		retryAfterStr := resp.Header.Get("Retry-After")
		retryAfterInt, err := strconv.Atoi(retryAfterStr)
		if err != nil {
			a.logger.Log.Errorln(err)
			retryAfterInt = retryAfterDefault
		}
		return &APIResponse{
			RetryAfter: retryAfterInt,
		}, errTooManyRequests
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
	a.logger.Log.Debug("apiwrapper returning response: ", response)

	return response, nil
}
