package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/maxzhirnov/rewardify/internal/models"
)

// MockRepo is a mock implementation of the repo interface
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockRepo) InsertNewUser(ctx context.Context, user models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func TestAuthService_Register(t *testing.T) {
	repo := new(MockRepo)
	authService := NewAuthService(repo, "secret")
	testCases := []struct {
		username    string
		password    string
		insertError error
		expectedErr error
	}{
		{
			username:    "testUser",
			password:    "password",
			insertError: nil,
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.username, func(t *testing.T) {
			repo.On("InsertNewUser", mock.Anything, mock.Anything).Return(testCase.insertError)
			err := authService.Register(context.Background(), testCase.username, testCase.password)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}

}

func TestAuthService_Authenticate(t *testing.T) {
	repo := new(MockRepo)

	authService := NewAuthService(repo, "secret")

	user := models.User{
		UUID:     "12345",
		Username: "testUser",
		Password: "$2y$10$FqkJ.hLVquuLIHX3quXjq.ikD7zQdfVGoHEUmiWy64v2Fv7Pr0V9.", // Hashed "password"
	}

	testCases := []struct {
		username    string
		password    string
		getUserErr  error
		expectedErr error
	}{
		{
			username:    "testUser",
			password:    "password",
			getUserErr:  nil,
			expectedErr: nil,
		},
		{
			username:    "nonExistentUser",
			password:    "password",
			getUserErr:  ErrInvalidCredentials,
			expectedErr: ErrInvalidCredentials,
		},
		{
			username:    "testUser",
			password:    "wrongPassword",
			getUserErr:  nil,
			expectedErr: ErrInvalidCredentials,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.username, func(t *testing.T) {
			// Mock the behavior of the repo's GetUserByUsername method
			repo.On("GetUserByUsername", mock.Anything, testCase.username).Return(user, testCase.getUserErr)

			// Call the Authenticate function
			_, err := authService.Authenticate(context.Background(), testCase.username, testCase.password)

			// Assert the expected error
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	repo := new(MockRepo)
	authService := NewAuthService(repo, "secret")
	mockToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6IkpvaG4gRG9lIiwidXVpZCI6IjEyMzQ1NiJ9.yt5p3if6g9R1oOwVHWpCSWD46UcxTFPx7WwwFjGPpkg"
	mockUser := &models.User{
		UUID:     "123456",
		Username: "John Doe",
	}
	user, err := authService.ValidateToken(mockToken)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, mockUser.Username, user.Username)
	assert.Equal(t, mockUser.UUID, user.UUID)

}
