package forms

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
	"github.com/naouuud/formulator-api/internal/models"
)

type Service interface {
	GetFormsByUserId(ctx context.Context, userId string) ([]models.FormSchema, error)
	CreateForm(ctx context.Context, userId string) (string, error)
	InitializeUserForms(ctx context.Context, userId string) ([]models.FormSchema, error)
	DeleteForm(ctx context.Context, id string) error
	UpdateFormSchema(ctx context.Context, id string, schema models.FormSchemaDB) error
}

type service struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) GetFormsByUserId(ctx context.Context, userID string) ([]models.FormSchema, error) {
	var schemas []models.FormSchema
	userIdPg := pgtype.Text{String: userID, Valid: true}
	forms, err := s.repo.GetFormsByUserId(ctx, userIdPg)
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

func (s *service) CreateForm(ctx context.Context, userID string) (string, error) {
	ID := uuid.New().String()
	userIdPg := pgtype.Text{String: userID, Valid: true}
	dbSchema := models.FormSchemaDB{Nodes: []models.Node{}}
	encoded, err := json.Marshal(dbSchema)
	if err != nil {
		logErr(err)
		return "", err
	}
	params := repo.CreateFormParams{
		ID:         ID,
		UserID:     userIdPg,
		FormSchema: encoded,
	}
	err = s.repo.CreateForm(ctx, params)
	if err != nil {
		slog.Error("Failed to create form", "err", err)
		return "", err
	}
	return ID, err
}

func (s *service) InitializeUserForms(ctx context.Context, userID string) ([]models.FormSchema, error) {
	ID, err := s.CreateForm(ctx, userID)
	schema := models.FormSchema{
		ID:    ID,
		Nodes: []models.Node{},
	}
	return []models.FormSchema{schema}, err
}

func (s *service) DeleteForm(ctx context.Context, ID string) error {
	err := s.repo.DeleteForm(ctx, ID)
	if err != nil {
		slog.Error("delete form failed", "err", err)
	}
	return err
}

func (s *service) UpdateFormSchema(ctx context.Context, id string, schema models.FormSchemaDB) error {
	encoded, err := json.Marshal(schema); 
	if err != nil {
		slog.Error("failed to encode form schema", "err", err)
		return err
	}
	params := repo.UpdateFormSchemaParams{
		FormSchema: encoded,
		ID: id,
	}
	if err := s.repo.UpdateFormSchema(ctx, params); err != nil {
		slog.Error("failed to update form schema", "err", err)
		return err
	}
	return nil
}

func logErr(err error) {
	log.Printf("Auth service error: %s", err.Error())
}
