package users

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
)

var ErrNotFound = errors.New("User not found")
var ErrUserNotCreated = errors.New("User not created")

type Service interface {
	GetUserById(ctx context.Context, id string) (repo.User, error)
	CreateUser(ctx context.Context, userDto CreateUserDto) error
}

type ServiceCt struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &ServiceCt{
		repo: repo,
	}
}

func (this *ServiceCt) GetUserById(ctx context.Context, id string) (repo.User, error) {
	user, err := this.repo.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.User{}, ErrNotFound
		}
		// handle other errors
	}
	return user, err
}

func (this *ServiceCt) CreateUser(ctx context.Context, userDto CreateUserDto) error {
	id := uuid.New().String()
	userParams := repo.CreateUserParams{
		ID: id,
		Username: userDto.Username,
		FirstName:userDto.FirstName,
		LastName: userDto.LastName,
	}
	err := this.repo.CreateUser(ctx, userParams) 
	if err != nil {
		return ErrUserNotCreated
	}
	log.Printf("User created id = %s", id)
	return err
}