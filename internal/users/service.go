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
var ErrUsernameExists = errors.New("Username already exists")

type Service interface {
	GetUserById(ctx context.Context, id string) (repo.User, error)
	CreateUser(ctx context.Context, userDto CreateUserDto) error
	CreateAnonUser(ctx context.Context) error
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

func (this *ServiceCt) UsernameExists(ctx context.Context, username string) (int64, error) {
	count, err := this.repo.UsernameExists(ctx, username)
	if (err != nil) {
		logErr(err)
		return -1, err
	}
	return count, err
}

func (this *ServiceCt) CreateUser(ctx context.Context, userDto CreateUserDto) error {
	count, err := this.UsernameExists(ctx, userDto.Username)
	if (err != nil) {
		logErr(err)
		return err
	}
	if count > 0 {
		log.Println("Account creation failed, username " + userDto.Username + " already exists")
		return ErrUsernameExists
	}
	id := uuid.New().String()
	userParams := repo.CreateUserParams{
		ID:        id,
		Username:  userDto.Username,
		FirstName: userDto.FirstName,
		LastName:  userDto.LastName,
	}
	err = this.repo.CreateUser(ctx, userParams)
	if err != nil {
		logErr(err)
		return ErrUserNotCreated
	}
	log.Printf("User created with id = %s", id)
	return err
}

func (this *ServiceCt) CreateAnonUser(ctx context.Context) error {
	id := uuid.New().String()
	// Generate random int
	userParams := repo.CreateUserParams{
		ID:        id,
		Username:  "anon", // append random int
		FirstName: "",
		LastName:  "",
	}
	err := this.repo.CreateUser(ctx, userParams)
	if err != nil {
		logErr(err)
		return ErrUserNotCreated
	}
	log.Printf("Anon user created with id = %s", id)
	return err
}

func logErr(err error) {
	log.Printf("Service error: %s", err.Error())
}
