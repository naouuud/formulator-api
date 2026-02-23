package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
	"github.com/naouuud/formulator-api/internal/auth"
	"github.com/naouuud/formulator-api/internal/forms"
	"github.com/naouuud/formulator-api/internal/users"
)

type App struct {
	config AppConfig
	conn   *pgx.Conn
}

type AppConfig struct {
	port string
	dsn  string
}

func (this *App) run(h http.Handler) error {
	server := http.Server{
		Addr:    this.config.port,
		Handler: h,
	}
	log.Printf("Server has started at address %s", this.config.port)
	return server.ListenAndServe()
}

// func handlePreflight(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method == http.MethodOptions {
// 			w.WriteHeader(200)
// 		}
// 	})
// }

func (this *App) mount() http.Handler {
	router := chi.NewRouter()
	
	// Initialize repo
	repo := repo.New(this.conn)
	// Initialize services 
	authSvc := auth.NewService(repo)
	userSvc := users.NewService(repo)
	formSvc := forms.NewService(repo)

	// Middlewares
	router.Use(middleware.Logger)
	router.Use(middleware.SetHeader("Access-Control-Allow-Origin", "http://localhost:4200"))
	router.Use(middleware.SetHeader("Access-Control-Allow-Headers", "Authorization"))

	// Routes
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Health check!")
		w.Write([]byte("All good!"))
	})
	// Auth
	authHandler := auth.NewHandler(authSvc, userSvc, formSvc)
	router.Route("/auth", func(r chi.Router) {
		r.Options("/me", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})
		r.Get("/me", http.HandlerFunc(authHandler.Bootstrap))
	})
	// User
	userHandler := users.NewHandler(userSvc)
	router.Route("/user", func(r chi.Router) {
		r.Get("/{id}", userHandler.GetUserById)
		r.Post("/create", userHandler.CreateUser)
		// r.Post("/createanon", userHandler.CreateAnonUser)
	})
	// Forms
	formsHandler := forms.NewHandler(formSvc)
	router.Route("/form", func(r chi.Router) {
		r.Get("/", formsHandler.CreateForm)
		r.Post("/", formsHandler.CreateForm)
		r.Put("/", formsHandler.UpdateFormSchema)
		r.Delete("/", formsHandler.DeleteForm)
	})
	return router
}

// func myHandler(h http.Handler) http.Handler {
// 	log.Printf("Handler value: %v", h)
// 	return h
// }
