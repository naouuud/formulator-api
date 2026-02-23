package auth

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
)

var ErrTokenParse = errors.New("Error parsing token")
var ErrInvalidToken = errors.New("Invalid token")
var ErrUserNotFound = errors.New("User not found")
var ErrUserNotCreated = errors.New("Failed to create user")

type Service interface {
	GenerateToken(userID string) (string, error)
	ValidateToken(ctx context.Context, tokenStr string) (string, error)
}

type service struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &service{
		repo: repo,
	}
}

var jwtKey = []byte(os.Getenv("SECRET"))

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *service) GenerateToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(365 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func (s *service) ValidateToken(ctx context.Context, tokenStr string) (string, error) {
	var user repo.User
	// Parse token
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		slog.Error("Failed to parse token", "err", err)
		return user.ID, ErrTokenParse
	}
	// Validate token and check claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		slog.Error("Invalid token received", "err", err)
		return user.ID, ErrInvalidToken
	}
	// Check user exists
	if user, err = s.repo.GetUserById(ctx, claims.UserID); err != nil {
		slog.Error("Failed to fetch user", "err", err)
		return user.ID, ErrUserNotFound
	}
	return user.ID, err
}
