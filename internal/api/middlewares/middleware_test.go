package middlewares

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/maxzhirnov/rewardify/internal/auth"
	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateToken(token string) (models.User, error) {
	args := m.Called(token)
	return args.Get(0).(models.User), args.Error(1)
}

// Test GzipMiddleware
func TestGzipMiddleware(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Test response"))
	})

	req.Header.Set("Accept-Encoding", "gzip")

	middleware := GzipMiddleware(handler)
	middleware.ServeHTTP(rr, req)

	reader, _ := gzip.NewReader(rr.Body)
	defer reader.Close()
	decompressed, _ := io.ReadAll(reader)

	assert.Equal(t, "Test response", string(decompressed))
	assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))
}

// Test AuthMiddleware
func TestAuthMiddleware(t *testing.T) {
	l, _ := logger.NewLogger(logger.ErrorLevel, false)
	mockAuthService := new(MockAuthService)
	middlewareStruct := NewMiddlewares(mockAuthService, l)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Test valid token
	user := models.User{Username: "test", UUID: "uuid"}
	mockAuthService.On("ValidateToken", "valid-token").Return(user, nil)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: auth.JWTCookeName, Value: "valid-token"})
	rr := httptest.NewRecorder()

	middlewareStruct.AuthMiddleware(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())

	// Test unauthorized (invalid token)
	mockAuthService.On("ValidateToken", "invalid-token").Return(models.User{}, errors.New("unauthorized"))

	req = httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: auth.JWTCookeName, Value: "invalid-token"})
	rr = httptest.NewRecorder()

	middlewareStruct.AuthMiddleware(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Equal(t, "unauthorized", rr.Body.String())

	// Test unauthorized (no token)
	req = httptest.NewRequest("GET", "/", nil)
	rr = httptest.NewRecorder()

	middlewareStruct.AuthMiddleware(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Equal(t, "unauthorized", rr.Body.String())
}
