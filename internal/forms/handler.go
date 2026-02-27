package forms

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/naouuud/formulator-api/internal/models"
)

type CreateFormRes struct {
	Status string `json:"status"`
	ID     string `json:"id"`
}

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) CreateForm(w http.ResponseWriter, r *http.Request) {
	// Authorize
	userID := r.Context().Value("userID")
	if userID == nil {
		http.Error(w, "forbidden resource", http.StatusForbidden)
		return
	}
	userIDStr := userID.(string)
	formID, err := h.service.CreateForm(r.Context(), userIDStr)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateFormRes{
		Status: "ok",
		ID:     formID,
	})
}

func (h *handler) DeleteForm(w http.ResponseWriter, r *http.Request) {
	// Extract authenticated user
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	// Extract formID
	formID := chi.URLParam(r, "id")
	if formID == "" {
		slog.Error("empty formID with DELETE request")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	// Fetch form
	form, err := h.service.GetFormById(r.Context(), formID)
	if err != nil {
		if errors.Is(err, ErrFormNotFound) {
			http.Error(w, "form not found", http.StatusNotFound)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	// Ownership validation
	if form.UserID.String != userID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return 
	}	
	// Perform delete
	if err := h.service.DeleteForm(r.Context(), formID); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) UpdateFormSchema(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var body models.FormSchema
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.Error("failed to parse request body", "err", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	form, err := h.service.GetFormById(r.Context(), body.ID)
	if err != nil {
		if errors.Is(err, ErrFormNotFound) {
			http.Error(w, "form not found", http.StatusNotFound)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if form.UserID.String != userID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return 
	}	
	dbSchema := models.FormSchemaDB{
		Title: body.Title,
		Nodes: body.Nodes,
	}
	if err := h.service.UpdateFormSchema(r.Context(), body.ID, dbSchema); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
