package users

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (this *handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	user, err := this.service.GetUserById(r.Context(), idStr)
	if err != nil {
		httpError(w, err)
		return
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

type CreateUserDto struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (this *handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var dto CreateUserDto
	err := json.NewDecoder(r.Body).Decode(&dto)
	if err != nil {
		logErr(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// log.Printf("createUserDto = %+v", dto)

	err = this.service.CreateUser(r.Context(), dto)
	if err != nil {
		httpError(w, err)
		return
	}
}

// func (this *handler) CreateAnonUser(w http.ResponseWriter, r *http.Request) {
// 	err := this.service.createAnonUser(r.Context())
// 	if err != nil {
// 		httpError(w, err)
// 		return
// 	}
// }

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
