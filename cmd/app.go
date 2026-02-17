package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
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
	router.Get("/health", func(w http.ResponseWriter, r *http.Request ) {
		log.Printf("Health check!")
		w.Write([]byte("All good!"))
	})
	return router
}