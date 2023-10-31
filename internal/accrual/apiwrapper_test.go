package accrual

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
)

// API mock
func MockServer(handleFunc http.HandlerFunc) *httptest.Server {
	r := chi.NewRouter()
	r.Get("/api/orders/{orderID}", handleFunc)
	return httptest.NewServer(r)
}

func Test_ApiWrapper_Fetch(t *testing.T) {
	// Creating logger as it's accrual service dependency
	l, _ := logger.NewLogger(logger.ErrorLevel, false)

	tests := []struct {
		name           string
		inputOrderID   string
		serverResponse http.HandlerFunc
		expected       *APIResponse
		expectedErr    error
	}{
		{
			name:         "happy path",
			inputOrderID: "123456",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				orderID := chi.URLParam(r, "orderID")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf(`
{
"order": "%s",
"status": "PROCESSED",
"accrual": 732.3
}`, orderID)))
			},
			expected: &APIResponse{
				Order:   "123456",
				Status:  models.BonusAccrualStatusProcessed.String(),
				Accrual: 732.3,
			},
			expectedErr: nil,
		},
		{
			name:         "Handling 204",
			inputOrderID: "123456",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			expected:    nil,
			expectedErr: errNotRegistered,
		},
		{
			name:         "Handling 429",
			inputOrderID: "123456",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
			},
			expected:    &APIResponse{Order: "", Status: "", Accrual: 0, RetryAfter: 30},
			expectedErr: errTooManyRequests,
		},
		{
			name:         "Handling 500",
			inputOrderID: "123456",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expected:    nil,
			expectedErr: errBadRequest,
		},
		{
			name:         "Handling 409",
			inputOrderID: "123456",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusConflict)
			},
			expected:    nil,
			expectedErr: errBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpClient := &http.Client{}
			server := MockServer(tt.serverResponse)
			defer server.Close()
			apiWrapper := NewAPIWrapper(server.URL, httpClient, l)
			result, err := apiWrapper.Fetch(context.TODO(), tt.inputOrderID)

			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
