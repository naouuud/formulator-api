package users

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"math/rand"
	"strconv"

	"github.com/google/uuid"
	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
)

var ErrNotFound = errors.New("User not found")
var ErrUserNotCreated = errors.New("Failed to create user")
var ErrUsernameExists = errors.New("Username already exists")

type Service interface {
	GetUserById(ctx context.Context, id string) (repo.User, error)
	CreateUser(ctx context.Context, userDto CreateUserDto) error
	CreateAnonUser(ctx context.Context) error
}

type serviceCt struct {
	repo repo.Querier
}

func NewServiceCt(repo repo.Querier) Service {
	return &serviceCt{
		repo: repo,
	}
}

func (this *serviceCt) GetUserById(ctx context.Context, id string) (repo.User, error) {
	user, err := this.repo.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.User{}, ErrNotFound
		}
		logErr(err)
		return repo.User{}, err
	}
	return user, err
}

func (this *serviceCt) UsernameExists(ctx context.Context, username string) (int64, error) {
	count, err := this.repo.UsernameExists(ctx, username)
	if err != nil {
		logErr(err)
		return -1, err
	}
	return count, err
}

func (this *serviceCt) CreateUser(ctx context.Context, userDto CreateUserDto) error {
	count, err := this.UsernameExists(ctx, userDto.Username)
	if err != nil {
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

func (this *serviceCt) CreateAnonUser(ctx context.Context) error {
	id := uuid.New().String()
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
		return err
	}
	for count > 0 {
		randInt = src.Int63()
		userParams.Username = "anon-" + strconv.FormatInt(randInt, 10)
		count, err = this.UsernameExists(ctx, userParams.Username)
		if err != nil {
			logErr(err)
			return err
		}
	}
	err = this.repo.CreateUser(ctx, userParams)
	if err != nil {
		logErr(err)
		return ErrUserNotCreated
	}
	log.Printf("Anon user created with id = %s", id)
	return err
}

func logErr(err error) {
	log.Printf("User service error: %s", err.Error())
}

