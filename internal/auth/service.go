package auth

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
)

var ErrTokenParse = errors.New("Error parsing token")
var ErrInvalidToken = errors.New("Invalid token")
var ErrNotFound = errors.New("User not found")
var ErrUserNotCreated = errors.New("Failed to create user")

type Service interface {
	GenerateToken(userID string) (string, error)
	ValidateToken(ctx context.Context, tokenStr string) (*Claims, error)
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

func (this *service) GenerateToken(userID string) (string, error) {
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

func (this *service) ValidateToken(ctx context.Context, tokenStr string) (*Claims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		logErr(err)
		return nil, ErrTokenParse
	}
	// Validate token and check claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		log.Println("Invalid token received")
		return nil, ErrInvalidToken
	}
	return claims, err
}

func logErr(err error) {
	log.Printf("Auth service error: %s", err.Error())
}
