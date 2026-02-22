package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
	"github.com/naouuud/formulator-api/internal/models"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

type AuthResp struct {
	Auth   string              `json:"auth"`
	Forms  []models.FormSchema `json:"forms"`
	Status string              `json:"status"`
}

func (this *handler) Bootstrap(w http.ResponseWriter, r *http.Request) {
	tokenStr := extractToken(r)
	var user *repo.User
	var err error
	var idStr string
	var forms []models.FormSchema
	var status string
	// Validate token and check for existing user
	// Skip if token string empty
	if tokenStr != "" {
		user, err = this.service.ValidateToken(r.Context(), tokenStr)
		// Continue if token valid and user exists
		if user != nil && err == nil {
			idStr = user.ID
			status = "returning"
			forms, err = this.service.GetFormsByUserId(r.Context(), idStr)
			// If form fetch fails for verified user send 500, do not overwrite data
			if err != nil {
				httpError(w, err)
				return
			}
		}
	}
	// If no token, token valiation fails or user not found create Anon user and issue new token
	if idStr == "" {
		// Create anon user
		status = "new"
		idStr, err = this.service.CreateAnonUser(r.Context())
		if err != nil {
			httpError(w, err)
			return
		}
		// Create new token
		tokenStr, err = this.service.GenerateToken(idStr)
		if err != nil {
			log.Println("Failed to generate token")
			httpError(w, err)
			return
		}
		// Create first form
		forms, err = this.service.CreateFirstForm(r.Context(), idStr)
		if err != nil {
			httpError(w, err)
			return
		}
		// Optional set cookie
	}
	// Response
	resp := AuthResp{
		Auth:   tokenStr,
		Forms:  forms,
		Status: status,
	}
	// log.Printf("%+v", resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func extractToken(r *http.Request) string {
	// 1. Check Authorization header
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}

	// 2. Check cookie
	cookie, err := r.Cookie("token")
	if err == nil {
		return cookie.Value
	}

	return ""
}

func httpError(w http.ResponseWriter, err error) {
	switch err {
	case ErrUserNotCreated:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	case ErrNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, "Server error", http.StatusInternalServerError)
	}
}
