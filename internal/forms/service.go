package forms

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
	"github.com/naouuud/formulator-api/internal/models"
)

type Service interface {
	GetFormsByUserId(ctx context.Context, userId string) ([]models.FormSchema, error)
	CreateForm(ctx context.Context, userId string) error
	CreateEmptyFormSchemaList(ctx context.Context, userId string) ([]models.FormSchema, error)
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

func (this *service) GetFormsByUserId(ctx context.Context, userId string) ([]models.FormSchema, error) {
	var schemas []models.FormSchema
	userIdPg := pgtype.Text{String: userId, Valid: true}
	forms, err := this.repo.GetFormsByUserId(ctx, userIdPg)
	if err != nil {
		logErr(err)
		return schemas, err
	}
	for _, v := range forms {
		dbSchema := models.FormSchemaDB{}
		err = json.Unmarshal(v.FormSchema, &dbSchema)
		if err != nil {
			logErr(err)
			return schemas, err
		}
		schema := models.FormSchema{
			ID:    v.ID,
			Title: dbSchema.Title,
			Nodes: dbSchema.Nodes,
		}
		schemas = append(schemas, schema)
	}
	return schemas, err
}

func (this *service) CreateForm(ctx context.Context, userId string) error {
	return nil
}

func (this *service) CreateEmptyFormSchemaList(ctx context.Context, userId string) ([]models.FormSchema, error) {
	id := uuid.New().String()
	schema := models.FormSchema{ID: id, Nodes: []any{}}
	dbSchema := models.FormSchemaDB{Nodes: []any{}}
	encoded, err := json.Marshal(dbSchema)
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
	// log.Printf("First form created for new user, id = %s", id)
	return []models.FormSchema{schema}, err
}

func (this *service) DeleteForm(ctx context.Context, id string) error {
	return nil
}

func (this *service) UpdateFormSchema(ctx context.Context, id string, formSchema []byte) error {
	return nil
}

func logErr(err error) {
	log.Printf("Auth service error: %s", err.Error())
}
