package auth

import (
	"context"
	"errors"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/maxzhirnov/rewardify/internal/models"
	r "github.com/maxzhirnov/rewardify/internal/repo"
)

var (
	ErrUserAlreadyExist   = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type ContextCustomKey string

const (
	UsernameContextKey ContextCustomKey = "username"
	UUIDContextKey     ContextCustomKey = "uuid"
	JWTCookeName                        = "token"
)

type repo interface {
	GetUserByUsername(ctx context.Context, username string) (models.User, error)
	InsertNewUser(ctx context.Context, user models.User) error
}

type AuthService struct {
	repo      repo
	secretKey string
}

func NewAuthService(r repo, secretKey string) *AuthService {
	return &AuthService{
		repo:      r,
		secretKey: secretKey,
	}
}

func (a *AuthService) Register(ctx context.Context, username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := models.User{
		UUID:     uuid.New().String(),
		Username: username,
		Password: string(hashedPassword),
	}

	err = a.repo.InsertNewUser(ctx, user)
	if errors.Is(err, r.ErrUserAlreadyExist) {
		return ErrUserAlreadyExist
	}
	return err
}

func (a *AuthService) Authenticate(ctx context.Context, username, password string) (string, error) {
	user, err := a.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uuid":     user.UUID,
		"username": username,
	})

	tokenString, err := token.SignedString([]byte(a.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *AuthService) ValidateToken(tokenString string) (models.User, error) {
	var claims jwt.MapClaims
	user := models.User{}

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.secretKey), nil
	})

	if err != nil {
		return user, err
	}

	if !token.Valid {
		return user, ErrInvalidToken
	}

	user.Username = claims["username"].(string)
	user.UUID = claims["uuid"].(string)

	return user, nil
}
