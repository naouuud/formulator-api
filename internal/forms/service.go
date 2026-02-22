package forms

import (
	"context"

	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
	"github.com/naouuud/formulator-api/internal/models"
)

type Service interface {
	GetFormsByUserId(ctx context.Context, userId string) []models.FormSchema
	CreateForm(ctx context.Context, userId string) error
	DeleteForm(ctx context.Context, id string) error
	UpdateFormSchema(ctx context.Context, id string, formSchema []byte) error
}

type service struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &service{
		repo: repo,
	}
}

func (this *service) GetFormsByUserId(ctx context.Context, userId string) []models.FormSchema {
	return []models.FormSchema{}
}

func (this *service) CreateForm(ctx context.Context, userId string) error {
	return nil
}

func (this *service) DeleteForm(ctx context.Context, id string) error {
	return nil
}

func (this *service) UpdateFormSchema(ctx context.Context, id string, formSchema []byte) error {
	return nil
}



