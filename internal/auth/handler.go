package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/naouuud/formulator-api/internal/forms"
	"github.com/naouuud/formulator-api/internal/models"
	"github.com/naouuud/formulator-api/internal/users"
)

type handler struct {
	svc     Service
	userSvc users.Service
	formSvc forms.Service
}

func NewHandler(service Service, userSvc users.Service, formSvc forms.Service) *handler {
	return &handler{
		svc:     service,
		userSvc: userSvc,
		formSvc: formSvc,
	}
}

type AuthResp struct {
	ID     string              `json:"id"`
	Auth   string              `json:"auth"`
	Forms  []models.FormSchema `json:"forms"`
	Status string              `json:"status"`
}

func (h *handler) Bootstrap(w http.ResponseWriter, r *http.Request) {
	tokenStr := extractToken(r)
	var err error
	var userID string
	var status string
	var forms []models.FormSchema
	// Validate token and check for existing user
	// Skip if token string empty
	if tokenStr != "" {
		userID, err = h.svc.ValidateToken(r.Context(), tokenStr)
		// Continue if token valid & user exists
		if userID != "" && err == nil {
			status = "returning"
			forms, err = h.formSvc.GetFormsByUserId(r.Context(), userID)
			// If form fetch fails for verified user send 500, do not overwrite data
			if err != nil {
				httpError(w, err)
				return
			}
		}
	}
	// If no token, token valiation fails or user not found create Anon user and issue new token
	if userID == "" {
		// Create anon user
		status = "new"
		userID, err = h.userSvc.CreateAnonUser(r.Context())
		if err != nil {
			httpError(w, err)
			return
		}
		// Create new token
		tokenStr, err = h.svc.GenerateToken(userID)
		if err != nil {
			log.Println("Failed to generate token")
			httpError(w, err)
			return
		}
		// Create first form
		forms, err = h.formSvc.InitializeUserForms(r.Context(), userID)
		if err != nil {
			httpError(w, err)
			return
		}
		// Optional set cookie
	}
	// Response
	resp := AuthResp{
		ID:     userID,
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
	case ErrUserNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, "Server error", http.StatusInternalServerError)
	}
}
