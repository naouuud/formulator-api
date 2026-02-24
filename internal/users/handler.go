package users

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
)

type CreateUserReq struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

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
		if errors.Is(err, ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var body CreateUserReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.Error("failed to parse request body", "err", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	params := repo.CreateUserParams{
		Username:  body.Username,
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Role:      "registered",
	}
	if _, err := h.service.CreateUser(r.Context(), params); err != nil {
		if errors.Is(err, ErrUsernameExists) {
			http.Error(w, "username already exists", http.StatusConflict)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
}
