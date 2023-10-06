package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	app2 "github.com/maxzhirnov/rewardify/internal/app"
	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
)

// MockApp is a mock for app service
type MockApp struct {
	mock.Mock
}

func (m *MockApp) Register(ctx context.Context, username, password string) error {
	args := m.Called(ctx, username, password)
	return args.Error(0)
}
func (m *MockApp) Authenticate(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}
func (m *MockApp) UploadOrder(ctx context.Context, orderNumber, userUUID string) error {
	args := m.Called(ctx, orderNumber, userUUID)
	return args.Error(0)
}
func (m *MockApp) GetAllOrders(ctx context.Context, userUUID string) ([]models.Order, error) {
	args := m.Called(ctx, userUUID)
	return args.Get(0).([]models.Order), args.Error(1)
}
func (m *MockApp) GetBalance(ctx context.Context, userUUID string) (models.UsersBalance, error) {
	args := m.Called(ctx, userUUID)
	return args.Get(0).(models.UsersBalance), args.Error(1)
}
func (m *MockApp) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestHandlers_HandleRegister(t *testing.T) {
	r := chi.NewRouter()
	app := new(MockApp)
	app.On("Register", mock.Anything, "max", "123456").Return(nil)
	app.On("Register", mock.Anything, "token_problem_user", "123456").Return(nil)
	app.On("Register", mock.Anything, "existing", "123456").Return(app2.ErrUserAlreadyExist)
	app.On("Register", mock.Anything, "wrong_user", "123456").Return(errors.New("unexpected error"))
	app.On("Authenticate", mock.Anything, "max", "123456").Return("654321", nil)
	app.On("Authenticate", mock.Anything, "token_problem_user", "123456").Return("654321", errors.New("token problem"))
	l, err := logger.NewLogger(logger.ErrorLevel, false)
	if err != nil {
		t.Fatal(err)
	}
	handlers := NewHandlers(app, l)
	r.Post("/register", handlers.HandleRegister)

	type expected struct {
		statusCode       int
		shouldHaveCookie bool
		tokenCookieName  string
		tokenCookieVal   string
	}

	type reqParams struct {
		method string
		url    string
		body   io.Reader
	}
	tests := []struct {
		name     string
		req      reqParams
		expected expected
	}{
		{
			name: "happy path",
			req: reqParams{
				method: http.MethodPost,
				url:    "/register",
				body: bytes.NewReader([]byte(`{
    "login": "max",
    "password": "123456"
} `)),
			},
			expected: expected{
				statusCode:       200,
				shouldHaveCookie: true,
				tokenCookieName:  "token",
				tokenCookieVal:   "654321",
			},
		},
		{
			name: "bad json",
			req: reqParams{
				method: http.MethodPost,
				url:    "/register",
				body: bytes.NewReader([]byte(`{
    "login": "max"
    "password": "123456"
} `)),
			},
			expected: expected{
				statusCode:       400,
				shouldHaveCookie: false,
				tokenCookieName:  "",
				tokenCookieVal:   "",
			},
		},
		{
			name: "not enough body params",
			req: reqParams{
				method: http.MethodPost,
				url:    "/register",
				body: bytes.NewReader([]byte(`{
    "login": "max"
} `)),
			},
			expected: expected{
				statusCode:       400,
				shouldHaveCookie: false,
				tokenCookieName:  "",
				tokenCookieVal:   "",
			},
		},
		{
			name: "nil body",
			req: reqParams{
				method: http.MethodPost,
				url:    "/register",
				body:   nil,
			},
			expected: expected{
				statusCode:       400,
				shouldHaveCookie: false,
				tokenCookieName:  "",
				tokenCookieVal:   "",
			},
		},
		{
			name: "user already exists",
			req: reqParams{
				method: http.MethodPost,
				url:    "/register",
				body: bytes.NewReader([]byte(`{
    "login": "existing",
	"password": "123456"
} `)),
			},
			expected: expected{
				statusCode:       409,
				shouldHaveCookie: false,
				tokenCookieName:  "",
				tokenCookieVal:   "",
			},
		},
		{
			name: "user already exists",
			req: reqParams{
				method: http.MethodPost,
				url:    "/register",
				body: bytes.NewReader([]byte(`{
    "login": "wrong_user",
	"password": "123456"
} `)),
			},
			expected: expected{
				statusCode:       500,
				shouldHaveCookie: false,
				tokenCookieName:  "",
				tokenCookieVal:   "",
			},
		},
		{
			name: "user already exists",
			req: reqParams{
				method: http.MethodPost,
				url:    "/register",
				body: bytes.NewReader([]byte(`{
    "login": "token_problem_user",
	"password": "123456"
} `)),
			},
			expected: expected{
				statusCode:       500,
				shouldHaveCookie: false,
				tokenCookieName:  "",
				tokenCookieVal:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.req.method, tt.req.url, tt.req.body)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)
			assert.Equal(t, tt.expected.statusCode, rr.Code)
			defer rr.Result().Body.Close()
			cookies := rr.Result().Cookies()
			if tt.expected.shouldHaveCookie {
				var found bool
				for _, cookie := range cookies {
					if cookie.Name == tt.expected.tokenCookieName {
						found = true
						assert.Equal(t, tt.expected.tokenCookieVal, cookie.Value)
						break
					}
				}

				if !found {
					t.Errorf("Cookie \"%s\" not found", tt.expected.tokenCookieName)
				}
			} else {
				assert.Equal(t, 0, len(cookies))
			}
		})
	}
}
