package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/naouuud/formulator-api/internal/adapters/postgres/repo"
	"github.com/naouuud/formulator-api/internal/users"
)


type App struct {
	config AppConfig
	conn *pgx.Conn
}

type AppConfig struct {
	port string
	dsn string
}

func (this *App) run(h http.Handler) error {
	server := http.Server{
		Addr: this.config.port,
		Handler: h,
	}
	log.Printf("Server has started at address %s", this.config.port)
	return server.ListenAndServe()
}

func (this *App) mount() http.Handler {
	router := chi.NewRouter()

	// Middleware
	// Routes
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Health check!")
		w.Write([]byte("All good!"))
	})

	// User
	userHandler := users.NewHandler(users.NewService(repo.New(this.conn)))
	router.Route("/user", func(r chi.Router) {
		r.Get("/{id}", userHandler.GetUserById)
		r.Post("/create", userHandler.CreateUser)
		r.Post("/createanon", userHandler.CreateAnonUser)
	})
	return router
}

// func myHandler(h http.Handler) http.Handler {
// 	log.Printf("Handler value: %v", h)
// 	return h
// }
