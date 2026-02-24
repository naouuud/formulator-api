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

var ErrUserNotFound = errors.New("user not found")
var ErrUserNotCreated = errors.New("failed to create user")
var ErrUsernameExists = errors.New("username already exists")

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
			return user, ErrUserNotFound
		}
		slog.Error("failed to fetch user", "err", err)
	}
	return user, err
}

func (s *service) CreateUser(ctx context.Context, params repo.CreateUserParams) (ID string, err error) {
	var count int64
	if count, err = s.usernameExists(ctx, params.Username); err != nil {
		return
	}
	if count > 0 {
		slog.Info("failed to create account, duplicate username")
		err = ErrUsernameExists
		return
	}
	ID = uuid.New().String()
	userParams := repo.CreateUserParams{
		ID:        ID,
		Username:  params.Username,
		FirstName: params.FirstName,
		LastName:  params.LastName,
		Role:      params.Role,
	}
	if err = s.repo.CreateUser(ctx, userParams); err != nil {
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
		Username: "anon-" + strconv.FormatInt(randInt, 10),
		Role:     "anonymous",
	}
	ID, err := s.CreateUser(ctx, params)
	for err != nil {
		if errors.Is(err, ErrUsernameExists) {
			randInt = src.Int63()
			params.Username = "anon-" + strconv.FormatInt(randInt, 10)
			ID, err = s.CreateUser(ctx, params)
		} else {
			return "", err
		}
	}
	return ID, err
}

func (s *service) usernameExists(ctx context.Context, username string) (int64, error) {
	count, err := s.repo.UsernameExists(ctx, username)
	if err != nil {
		slog.Error("failed to check username", "err", err)
		return -1, err
	}
	return count, err
}
