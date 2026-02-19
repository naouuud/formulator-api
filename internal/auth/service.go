package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
)

var ErrTokenParse = errors.New("Error parsing token")
var ErrInvalidToken = errors.New("Invalid token")
var ErrNotFound = errors.New("User not found")
var ErrUserNotCreated = errors.New("Failed to create user")

type Service interface {
	GenerateToken(userID string) (string, error)
	ValidateToken(ctx context.Context, tokenStr string) (*repo.User, error)
	CreateAnonUser(ctx context.Context) (id string, err error)
}

type ServiceCt struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &ServiceCt{
		repo: repo,
	}
}

var jwtKey = []byte(os.Getenv("SECRET"))

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func (this *ServiceCt) GenerateToken(userID string) (string, error) {
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

func (this *ServiceCt) ValidateToken(ctx context.Context, tokenStr string) (*repo.User, error) {
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
	// Retrieve User
	return this.GetUserById(ctx, claims.UserID)
}

func (this *ServiceCt) GetUserById(ctx context.Context, id string) (*repo.User, error) {
	user, err := this.repo.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		logErr(err)
		return nil, err
	}
	return &user, err
}

func (this *ServiceCt) CreateAnonUser(ctx context.Context) (id string, err error) {
	id = uuid.New().String()
	// Generate random int
	src := rand.NewSource(100)
	randInt := src.Int63()
	userParams := repo.CreateUserParams{
		ID:        id,
		Username:  "anon-" + strconv.FormatInt(randInt, 10),
		FirstName: "",
		LastName:  "",
	}
	count, err := this.UsernameExists(ctx, userParams.Username)
	if err != nil {
		logErr(err)
		return id, err
	}
	for count > 0 {
		randInt = src.Int63()
		userParams.Username = "anon-" + strconv.FormatInt(randInt, 10)
		count, err = this.UsernameExists(ctx, userParams.Username)
		if err != nil {
			logErr(err)
			return id, err
		}
	}
	err = this.repo.CreateUser(ctx, userParams)
	if err != nil {
		logErr(err)
		return id, ErrUserNotCreated
	}
	return id, err
}

func (this *ServiceCt) UsernameExists(ctx context.Context, username string) (int64, error) {
	count, err := this.repo.UsernameExists(ctx, username)
	if err != nil {
		logErr(err)
		return -1, err
	}
	return count, err
}

func logErr(err error) {
	log.Printf("Auth service error: %s", err.Error())
}
