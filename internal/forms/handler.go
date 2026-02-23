package forms

import "net/http"

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (this *handler) GetFormsByUserId(w http.ResponseWriter, r *http.Request) {

}
func (this *handler) CreateForm(w http.ResponseWriter, r *http.Request) {

}
func (this *handler) DeleteForm(w http.ResponseWriter, r *http.Request)       {}
func (this *handler) UpdateFormSchema(w http.ResponseWriter, r *http.Request) {}
