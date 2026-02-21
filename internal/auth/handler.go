package auth

import (
	"encoding/json"
	"log"
	"net/http"

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

type AuthResp struct {
	Id    string       `json:"id"`
	Token string       `json:"token"`
	Forms []FormSchema `json:"forms"`
}

func (this *handler) Bootstrap(w http.ResponseWriter, r *http.Request) {
	tokenStr := extractToken(r)
	var user *repo.User
	var err error
	var idStr string
	var forms []FormSchema
	// Validate token and check for existing user
	if tokenStr != "" {
		user, err = this.service.ValidateToken(r.Context(), tokenStr)
		if user != nil && err == nil {
			idStr = user.ID
			forms, err = this.service.GetUserForms(r.Context(), idStr)
			if err != nil {
				httpError(w, err)
				return
			}
		}
	}
	// If anything has failed (idStr still == "") create Anon user
	if idStr == "" {
		// Create anon user
		idStr, err = this.service.CreateAnonUser(r.Context())
		if err != nil {
			httpError(w, err)
			return
		}
		log.Printf("Anon user created id = %s", idStr)
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
		Id:    idStr,
		Token: tokenStr,
		Forms: forms,
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
