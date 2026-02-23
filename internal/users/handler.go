package users

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	user, err := h.service.GetUserById(r.Context(), idStr)
	if err != nil {
		httpError(w, err)
		return
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var params repo.CreateUserParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		slog.Error("failed to parse request body", "err", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if _, err := h.service.CreateUser(r.Context(), params); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
}

func httpError(w http.ResponseWriter, err error) {
	switch err {
	case ErrUsernameExists:
		http.Error(w, err.Error(), http.StatusConflict)
	case ErrUserNotCreated:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	case ErrNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, "Server error", http.StatusInternalServerError)
	}
}
