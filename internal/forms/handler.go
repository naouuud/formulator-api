package forms

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	slog.Info("userID=", userID)
	if userID == nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	userIDStr := userID.(string)
	formID, err := h.service.CreateForm(r.Context(), userIDStr)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateFormRes{
		Status: "ok",
		ID:     formID,
	})
}

func (h *handler) DeleteForm(w http.ResponseWriter, r *http.Request) {
	// Authorize
	userID := r.Context().Value("userID")
	slog.Info("userID=", userID)
	if userID == nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	// Get FormID from url
	formID := chi.URLParam(r, "id")
	if formID == "" {
		slog.Error("empty formID with DELETE request")
		http.Error(w, "no formID provided", http.StatusBadRequest)
		return
	}
	if err := h.service.DeleteForm(r.Context(), formID); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) UpdateFormSchema(w http.ResponseWriter, r *http.Request) {}
