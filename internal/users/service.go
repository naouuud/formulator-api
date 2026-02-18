package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
)

var ErrNotFound = errors.New("User not found")

type Service interface {
	GetUserById(ctx context.Context, id string) (repo.User, error)
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