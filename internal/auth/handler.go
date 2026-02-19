package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

type AuthResp struct {
	Id    string `json:"id"`
	Token string `json:"token"`
	Role  string `json:"role"`
}

func (this *Handler) Bootstrap(w http.ResponseWriter, r *http.Request) {
	tokenStr := extractToken(r)
	var user *repo.User
	var err error
	var idStr = ""
	var roleStr = ""
	// Validate token and check for existing user
	if tokenStr != "" {
		user, err = this.service.ValidateToken(r.Context(), tokenStr)
		if user != nil && err == nil {
			idStr = user.ID
			roleStr = "registered"
		}
	}
	// If token validation or user fetch has failed
	if idStr == "" {
		// Create anon user
		idStr, err = this.service.CreateAnonUser(r.Context())
		if err != nil {
			httpError(w, err)
			return
		}
		roleStr = "anonymous"
		log.Printf("Anon user created id = %s", idStr)
		// Create new token
		tokenStr, err = this.service.GenerateToken(idStr)
		if err != nil {
			log.Println("Failed to generate token")
			http.Error(w, "Failed to create token", http.StatusInternalServerError)
			return
		}
		// Optional set cookie
	}
	// Response
	resp := AuthResp{
		Id:    idStr,
		Token: tokenStr,
		Role:  roleStr, // replace with User.Role column
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
