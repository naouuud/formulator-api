package users

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"math/rand"
	"strconv"

	"github.com/google/uuid"
	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
)

var ErrNotFound = errors.New("User not found")
var ErrUserNotCreated = errors.New("Failed to create user")
var ErrUsernameExists = errors.New("Username already exists")

type Service interface {
	GetUserById(ctx context.Context, ID string) (repo.User, error)
	CreateUser(ctx context.Context, params repo.CreateUserParams) (ID string, err error)
	CreateAnonUser(ctx context.Context) (string, error)
}

type service struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) GetUserById(ctx context.Context, id string) (repo.User, error) {
	user, err := s.repo.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logErr(err)
			return user, ErrNotFound
		}
		logErr(err)
		return user, err
	}
	return user, err
}

func (s *service) usernameExists(ctx context.Context, username string) (int64, error) {
	count, err := s.repo.UsernameExists(ctx, username)
	if err != nil {
		slog.Error("failed to check username", "err", err)
		return -1, err
	}
	return count, err
}

func (s *service) CreateUser(ctx context.Context, params repo.CreateUserParams) (ID string, err error) {
	count, err := s.usernameExists(ctx, params.Username)
	if err != nil {
		return
	}
	if count > 0 {
		log.Println("failed to create account, username " + params.Username + " already exists")
		err = ErrUsernameExists
		return
	}
	ID = uuid.New().String()
	// if Role == "", = "registered"
	userParams := repo.CreateUserParams{
		ID:        ID,
		Username:  params.Username,
		FirstName: params.FirstName,
		LastName:  params.LastName,
		Role: 	   params.Role,
	}
	err = s.repo.CreateUser(ctx, userParams)
	if err != nil {
		slog.Error("failed to create user", "err", err)
		err = ErrUserNotCreated
		return
	}
	log.Printf("user created with id = %s", ID)
	return
}

func (s *service) CreateAnonUser(ctx context.Context) (string, error) {
	src := rand.NewSource(100)
	randInt := src.Int63()
	params := repo.CreateUserParams{
		Username:  "anon-" + strconv.FormatInt(randInt, 10),
		Role:      "anonymous",
	}
	count, err := s.usernameExists(ctx, params.Username)
	if err != nil {
		slog.Error("failed to check username", "err", err)
		return "", err
	}
	for count > 0 {
		randInt = src.Int63()
		params.Username = "anon-" + strconv.FormatInt(randInt, 10)
		count, err = s.usernameExists(ctx, params.Username)
		if err != nil {
			slog.Error("failed to check username", "err", err)
			return "", err
		}
	}
	ID, err := s.CreateUser(ctx, params)
	if err != nil {
		return "", err
	}
	return ID, err
}

func logErr(err error) {
	log.Printf("User service error: %s", err.Error())
}
