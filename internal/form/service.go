package form

import "github.com/naouuud/formulator-api/internal/adapters/postgres/repo"

type Service interface {
}

type service struct {
	repo repo.Querier
}
