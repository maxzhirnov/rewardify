package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	app2 "github.com/maxzhirnov/rewardify/internal/app"
	"github.com/maxzhirnov/rewardify/internal/auth"
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
func (m *MockApp) CreateWithdrawal(ctx context.Context, userUUID string, orderNumber string, sum float32) error {
	args := m.Called(ctx, userUUID, orderNumber, sum)
	return args.Error(0)
}
func (m *MockApp) GetAllWithdrawals(ctx context.Context, usrUUID string) ([]models.Withdrawal, error) {
	args := m.Called(ctx, usrUUID)
	return args.Get(0).([]models.Withdrawal), args.Error(1)
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
				tokenCookieName:  auth.JWTCookeName,
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

			result := rr.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.expected.statusCode, rr.Code)
			cookies := result.Cookies()

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
func TestHandlers_HandleGetOrders(t *testing.T) {
	r := chi.NewRouter()
	app := new(MockApp)
	mockOrders := []models.Order{
		{
			OrderNumber:        "1",
			UserUUID:           "123456",
			BonusAccrualStatus: models.BonusAccrualStatusNew,
			BonusesAccrued:     152,
			CreatedAt:          time.Date(0001, 01, 01, 00, 00, 00, 00, time.UTC),
		},
		{
			OrderNumber:        "2",
			UserUUID:           "123456",
			BonusAccrualStatus: models.BonusAccrualStatusNew,
			BonusesAccrued:     15.5,
			CreatedAt:          time.Date(0001, 01, 01, 00, 00, 00, 00, time.UTC),
		},
	}
	app.On("GetAllOrders", mock.Anything, "123456").Return(mockOrders, nil)
	app.On("GetAllOrders", mock.Anything, "666").Return([]models.Order{}, errors.New("some error"))
	app.On("GetAllOrders", mock.Anything, "777").Return([]models.Order{}, nil)

	l, err := logger.NewLogger(logger.ErrorLevel, false)
	if err != nil {
		t.Fatal(err)
	}
	handlers := NewHandlers(app, l)
	r.Get("/get-orders", handlers.HandleGetOrders)

	type input struct {
		uuid string
	}

	type expected struct {
		statusCode int
		orders     []OrderDTO
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name:  "happy path",
			input: input{uuid: "123456"},
			expected: expected{
				statusCode: http.StatusOK,
				orders: []OrderDTO{
					{
						OrderNumber: "1",
						Status:      models.BonusAccrualStatusNew,
						Accrual:     152,
						UploadedAt:  "0001-01-01T00:00:00Z",
					},
					{
						OrderNumber: "2",
						Status:      models.BonusAccrualStatusNew,
						Accrual:     15.5,
						UploadedAt:  "0001-01-01T00:00:00Z",
					},
				},
			},
		},
		{
			name:  "empty uuid",
			input: input{uuid: ""},
			expected: expected{
				statusCode: http.StatusUnauthorized,
				orders:     []OrderDTO{},
			},
		},
		{
			name:  "app returns error",
			input: input{uuid: "666"},
			expected: expected{
				statusCode: http.StatusInternalServerError,
				orders:     []OrderDTO{},
			},
		},
		{
			name:  "empty orders",
			input: input{uuid: "777"},
			expected: expected{
				statusCode: http.StatusNoContent,
				orders:     []OrderDTO{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/get-orders", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.WithValue(req.Context(), auth.UUIDContextKey, tt.input.uuid)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req.WithContext(ctx))

			result := rr.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.expected.statusCode, rr.Code)

			if rr.Code == http.StatusOK {
				actualOrders := make([]OrderDTO, 0)
				err = json.NewDecoder(result.Body).Decode(&actualOrders)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, tt.expected.orders, actualOrders)
			}
		})
	}
}
func TestHandlers_HandleGetBalance(t *testing.T) {
	r := chi.NewRouter()
	app := new(MockApp)
	app.On("GetBalance", mock.Anything, "123456").Return(models.UsersBalance{
		UserUUID:  "123456",
		Earned:    1000,
		Current:   600,
		Withdrawn: 400,
	}, nil)

	app.On("GetBalance", mock.Anything, "666").Return(models.UsersBalance{}, errors.New("some error"))

	l, err := logger.NewLogger(logger.ErrorLevel, false)
	if err != nil {
		t.Fatal(err)
	}
	handlers := NewHandlers(app, l)
	r.Get("/get-balance", handlers.HandleGetBalance)

	type input struct {
		uuid string
	}

	type expected struct {
		statusCode int
		balance    GetBalanceResponseData
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name:  "happy path",
			input: input{uuid: "123456"},
			expected: expected{
				statusCode: http.StatusOK,
				balance: GetBalanceResponseData{
					Current:   600,
					Withdrawn: 400,
				},
			},
		},
		{
			name:  "empty uuid",
			input: input{uuid: ""},
			expected: expected{
				statusCode: http.StatusUnauthorized,
				balance:    GetBalanceResponseData{},
			},
		},
		{
			name:  "getting error from app",
			input: input{uuid: "666"},
			expected: expected{
				statusCode: http.StatusInternalServerError,
				balance:    GetBalanceResponseData{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/get-balance", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.WithValue(req.Context(), auth.UUIDContextKey, tt.input.uuid)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req.WithContext(ctx))

			result := rr.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.expected.statusCode, rr.Code)

			if rr.Code == http.StatusOK {
				actualBalance := GetBalanceResponseData{}
				err = json.NewDecoder(result.Body).Decode(&actualBalance)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, tt.expected.balance, actualBalance)
			}
		})
	}
}
func TestHandlers_HandleGetWithdrawals(t *testing.T) {
	r := chi.NewRouter()
	app := new(MockApp)
	mockWithdrawals := []models.Withdrawal{
		{
			OrderNumber: "1",
			UserUUID:    "123456",
			Amount:      100,
			CreatedAt:   time.Date(0001, 01, 01, 00, 00, 00, 00, time.UTC),
		},
		{
			OrderNumber: "2",
			UserUUID:    "123456",
			Amount:      152.2,
			CreatedAt:   time.Date(0001, 01, 01, 00, 00, 00, 00, time.UTC),
		},
	}
	app.On("GetAllWithdrawals", mock.Anything, "123456").Return(mockWithdrawals, nil)
	app.On("GetAllWithdrawals", mock.Anything, "666").Return([]models.Withdrawal{}, errors.New("some error"))
	app.On("GetAllWithdrawals", mock.Anything, "777").Return([]models.Withdrawal{}, nil)

	l, err := logger.NewLogger(logger.ErrorLevel, false)
	if err != nil {
		t.Fatal(err)
	}
	handlers := NewHandlers(app, l)
	r.Get("/get-withdrawals", handlers.HandleGetWithdrawals)

	type input struct {
		uuid string
	}

	type expected struct {
		statusCode int
		orders     []WithdrawalDTO
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name:  "happy path",
			input: input{uuid: "123456"},
			expected: expected{
				statusCode: http.StatusOK,
				orders: []WithdrawalDTO{
					{
						Order:       "1",
						Sum:         100,
						ProcessedAt: "0001-01-01T00:00:00Z",
					},
					{
						Order:       "2",
						Sum:         152.2,
						ProcessedAt: "0001-01-01T00:00:00Z",
					},
				},
			},
		},
		{
			name:  "empty uuid",
			input: input{uuid: ""},
			expected: expected{
				statusCode: http.StatusUnauthorized,
				orders:     []WithdrawalDTO{},
			},
		},
		{
			name:  "app returns error",
			input: input{uuid: "666"},
			expected: expected{
				statusCode: http.StatusInternalServerError,
				orders:     []WithdrawalDTO{},
			},
		},
		{
			name:  "empty orders",
			input: input{uuid: "777"},
			expected: expected{
				statusCode: http.StatusNoContent,
				orders:     []WithdrawalDTO{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/get-withdrawals", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.WithValue(req.Context(), auth.UUIDContextKey, tt.input.uuid)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req.WithContext(ctx))

			result := rr.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.expected.statusCode, rr.Code)

			if rr.Code == http.StatusOK {
				actualOrders := make([]WithdrawalDTO, 0)
				err = json.NewDecoder(result.Body).Decode(&actualOrders)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, tt.expected.orders, actualOrders)
			}
		})
	}
	app.AssertExpectations(t)
}
func TestHandlers_HandleLogin(t *testing.T) {
	r := chi.NewRouter()
	app := new(MockApp)

	app.On("Authenticate", mock.Anything, "max", "123456").Return("jwt", nil)
	app.On("Authenticate", mock.Anything, "wrong", "54321").Return("", errors.New("some error"))

	l, err := logger.NewLogger(logger.ErrorLevel, false)
	if err != nil {
		t.Fatal(err)
	}
	handlers := NewHandlers(app, l)
	r.Post("/login", handlers.HandleLogin)

	type input struct {
		data []byte
	}

	type expected struct {
		statusCode       int
		shouldHaveCookie bool
		jwtToken         string
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "happy path",
			input: input{
				data: []byte(`
{
	"login": "max",
	"password": "123456"
}`),
			},
			expected: expected{
				statusCode:       http.StatusOK,
				shouldHaveCookie: true,
				jwtToken:         "jwt",
			},
		},
		{
			name: "empty username and password",
			input: input{
				data: []byte(`
{
	"login": "",
	"password": ""
}`),
			},
			expected: expected{
				statusCode:       http.StatusBadRequest,
				shouldHaveCookie: false,
				jwtToken:         "",
			},
		},
		{
			name: "empty body",
			input: input{
				data: nil,
			},
			expected: expected{
				statusCode:       http.StatusBadRequest,
				shouldHaveCookie: false,
				jwtToken:         "",
			},
		},
		{
			name: "wrong creds",
			input: input{
				data: []byte(`
{
	"login": "wrong",
	"password": "54321"
}`),
			},
			expected: expected{
				statusCode:       http.StatusUnauthorized,
				shouldHaveCookie: false,
				jwtToken:         "jwt",
			},
		},
		{
			name: "bad json",
			input: input{
				data: []byte(`
{
	"login": "wrong
	"password": "54321"
}`),
			},
			expected: expected{
				statusCode:       http.StatusBadRequest,
				shouldHaveCookie: false,
				jwtToken:         "jwt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(tt.input.data))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			result := rr.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.expected.statusCode, rr.Code)

			if tt.expected.shouldHaveCookie {
				cookies := result.Cookies()
				var found bool
				for _, cookie := range cookies {
					if cookie.Name == auth.JWTCookeName {
						found = true
						assert.Equal(t, tt.expected.jwtToken, cookie.Value)
						break
					}
				}

				if !found {
					t.Errorf("Cookie \"%s\" not found", auth.JWTCookeName)
				}
			}

		})
	}
	app.AssertExpectations(t)
}
func TestHandlers_HandlePing(t *testing.T) {
	r := chi.NewRouter()
	app := new(MockApp)

	app.On("Ping", mock.Anything).Return(nil).Once()
	app.On("Ping", mock.Anything).Return(errors.New("some error"))

	l, err := logger.NewLogger(logger.ErrorLevel, false)
	if err != nil {
		t.Fatal(err)
	}
	handlers := NewHandlers(app, l)
	r.Get("/ping", handlers.HandlePing)

	tests := []struct {
		name               string
		expectedStatusCode int
	}{
		{
			name:               "happy path",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "app returns err",
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/ping", nil)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.WithValue(context.TODO(), auth.UsernameContextKey, "max")
			ctx = context.WithValue(ctx, auth.UUIDContextKey, "123456")

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req.WithContext(ctx))

			result := rr.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
		})
	}
	app.AssertExpectations(t)
}
func TestHandlers_HandleUploadOrder(t *testing.T) {
	r := chi.NewRouter()
	app := new(MockApp)

	app.On("UploadOrder", mock.Anything, "orderNumber", "uuid1").Return(nil)
	app.On("UploadOrder", mock.Anything, "invalidOrder", "uuid2").Return(app2.ErrInvalidOrderNumber)
	app.On("UploadOrder", mock.Anything, "existingOrder", "uuid3").Return(app2.ErrAlreadyCreatedByUser)
	app.On("UploadOrder", mock.Anything, "anotherUserOrder", "uuid4").Return(app2.ErrAlreadyCreatedByAnotherUser)
	app.On("UploadOrder", mock.Anything, "someError", "uuid5").Return(errors.New("some error"))

	l, err := logger.NewLogger(logger.ErrorLevel, false)
	if err != nil {
		t.Fatal(err)
	}
	handlers := NewHandlers(app, l)
	r.Post("/upload", handlers.HandleUploadOrder)

	type input struct {
		data []byte
		uuid string
	}

	type expected struct {
		statusCode   int
		responseText string
		err          error
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "happy path",
			input: input{
				data: []byte(`orderNumber`),
				uuid: "uuid1",
			},
			expected: expected{
				statusCode: http.StatusAccepted,
			},
		},
		{
			name: "app returns err invalid order",
			input: input{
				data: []byte(`invalidOrder`),
				uuid: "uuid2",
			},
			expected: expected{
				statusCode: http.StatusUnprocessableEntity,
			},
		},
		{
			name: "app returns err existing order",
			input: input{
				data: []byte(`existingOrder`),
				uuid: "uuid3",
			},
			expected: expected{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "app returns err another user's order",
			input: input{
				data: []byte(`anotherUserOrder`),
				uuid: "uuid4",
			},
			expected: expected{
				statusCode: http.StatusConflict,
			},
		},
		{
			name: "app returns err",
			input: input{
				data: []byte(`someError`),
				uuid: "uuid5",
			},
			expected: expected{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/upload", bytes.NewReader(tt.input.data))
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.WithValue(context.TODO(), auth.UUIDContextKey, tt.input.uuid)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req.WithContext(ctx))

			result := rr.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.expected.statusCode, rr.Code)
		})
	}
	app.AssertExpectations(t)
}
func TestHandlers_HandleWithdraw(t *testing.T) {
	r := chi.NewRouter()
	app := new(MockApp)

	app.On("CreateWithdrawal", mock.Anything, "uuid1", "order1", float32(100)).Return(nil)
	app.On("CreateWithdrawal", mock.Anything, "uuid2", "order2", float32(500)).Return(app2.ErrInsufficientFunds)
	app.On("CreateWithdrawal", mock.Anything, "uuid3", "order3", float32(100)).Return(app2.ErrInvalidOrderNumber)
	app.On("CreateWithdrawal", mock.Anything, "uuid4", "order4", float32(100)).Return(errors.New("some error"))

	l, err := logger.NewLogger(logger.ErrorLevel, false)
	if err != nil {
		t.Fatal(err)
	}
	handlers := NewHandlers(app, l)
	r.Post("/withdraw", handlers.HandleWithdraw)

	type input struct {
		uuid string
		data []byte
	}

	type expected struct {
		statusCode int
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "happy path",
			input: input{
				uuid: "uuid1",
				data: []byte(`
{
	"order": "order1",
	"sum": 100.0
}
`),
			},
			expected: expected{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "empty uuid",
			input: input{
				uuid: "",
				data: []byte(`
{
	"order": "order1",
	"sum": 100.0
}
`),
			},
			expected: expected{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "bad json",
			input: input{
				uuid: "uuid1",
				data: []byte(`
{
	"order": "order1"
	"sum": 100.0
}
`),
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "app returns err insufficient balance",
			input: input{
				uuid: "uuid2",
				data: []byte(`
{
	"order": "order2",
	"sum": 500.0
}
`),
			},
			expected: expected{
				statusCode: http.StatusPaymentRequired,
			},
		},
		{
			name: "app returns err ErrInvalidOrderNumber",
			input: input{
				uuid: "uuid3",
				data: []byte(`
{
	"order": "order3",
	"sum": 100.0
}
`),
			},
			expected: expected{
				statusCode: http.StatusUnprocessableEntity,
			},
		},
		{
			name: "app returns error",
			input: input{
				uuid: "uuid4",
				data: []byte(`
{
	"order": "order4",
	"sum": 100.0
}
`),
			},
			expected: expected{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/withdraw", bytes.NewReader(tt.input.data))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()

			ctx := context.WithValue(context.TODO(), auth.UUIDContextKey, tt.input.uuid)

			r.ServeHTTP(rr, req.WithContext(ctx))

			result := rr.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.expected.statusCode, rr.Code)
		})
	}
	app.AssertExpectations(t)
}
