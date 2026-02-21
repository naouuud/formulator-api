package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
)

var ErrTokenParse = errors.New("Error parsing token")
var ErrInvalidToken = errors.New("Invalid token")
var ErrNotFound = errors.New("User not found")
var ErrUserNotCreated = errors.New("Failed to create user")

type Service interface {
	GenerateToken(userID string) (string, error)
	ValidateToken(ctx context.Context, tokenStr string) (*repo.User, error)
	GetUserForms(ctx context.Context, userId string) ([]FormSchema, error)
	CreateAnonUser(ctx context.Context) (string, error)
	CreateFirstForm(ctx context.Context, userId string) ([]FormSchema, error)
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

type FormSchema struct {
	ID    string `json:"formId"`
	Title string `json:"formTitle"`
	Nodes []any  `json:"nodes"`
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

func (this *service) ValidateToken(ctx context.Context, tokenStr string) (*repo.User, error) {
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
	return this.getUserById(ctx, claims.UserID)
}

func (this *service) getUserById(ctx context.Context, id string) (*repo.User, error) {
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

func (this *service) GetUserForms(ctx context.Context, userId string) ([]FormSchema, error) {
	var schemas []FormSchema
	userIdPg := pgtype.Text{String: userId, Valid: true}
	forms, err := this.repo.GetFormsByUserId(ctx, userIdPg)
	if err != nil {
		logErr(err)
		return schemas, err
	}
	for _, v := range forms {
		schema := FormSchema{}
		err = json.Unmarshal(v.FormSchema, &schema)
		if err != nil {
			logErr(err)
			return schemas, err
		}
		schemas = append(schemas, schema)
	}
	return schemas, err
}

func (this *service) CreateAnonUser(ctx context.Context) (string, error) {
	id := uuid.New().String()
	// Generate random int
	src := rand.NewSource(100)
	randInt := src.Int63()
	userParams := repo.CreateUserParams{
		ID:        id,
		Username:  "anon-" + strconv.FormatInt(randInt, 10),
		FirstName: "",
		LastName:  "",
		Role:      "anonymous",
	}
	count, err := this.usernameExists(ctx, userParams.Username)
	if err != nil {
		logErr(err)
		return "", err
	}
	for count > 0 {
		randInt = src.Int63()
		userParams.Username = "anon-" + strconv.FormatInt(randInt, 10)
		count, err = this.usernameExists(ctx, userParams.Username)
		if err != nil {
			logErr(err)
			return "", err
		}
	}
	err = this.repo.CreateUser(ctx, userParams)
	if err != nil {
		logErr(err)
		return "", ErrUserNotCreated
	}
	return id, err
}

func (this *service) usernameExists(ctx context.Context, username string) (int64, error) {
	count, err := this.repo.UsernameExists(ctx, username)
	if err != nil {
		logErr(err)
		return -1, err
	}
	return count, err
}

func (this *service) CreateFirstForm(ctx context.Context, userId string) ([]FormSchema, error) {
	id := uuid.New().String()
	schema := FormSchema{ID: id, Nodes: []any{}}
	encoded, err := json.Marshal(schema)
	if err != nil {
		logErr(err)
		return nil, err
	}
	userIdPg := pgtype.Text{String: userId, Valid: true}
	params := repo.CreateFormParams{
		ID:         id,
		UserID:     userIdPg,
		FormSchema: encoded,
	}
	err = this.repo.CreateForm(ctx, params)
	if err != nil {
		logErr(err)
		return nil, err
	}
	log.Printf("First form created for user id = %s, form id = %s", userId, id)
	return []FormSchema{schema}, err
}

func logErr(err error) {
	log.Printf("Auth service error: %s", err.Error())
}
