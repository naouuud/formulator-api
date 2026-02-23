package main

import (
	"context"
	"log"
	"net/http"
	"strings"

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

// Middlware funcs
func handlePreflight(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200") // or *
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func AuthMiddleware(authSvc auth.Service) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
            authHeader := r.Header.Get("Authorization")
            if authHeader != "" {
            	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				if (tokenStr != "") {
            		userID, err := authSvc.ValidateToken(r.Context(), tokenStr)
					if (userID != "" && err == nil) {
					ctx = context.WithValue(ctx, "userID", userID)
					}
					// log.Printf("%+v", ctx)
				}
			}
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func (a *App) run(h http.Handler) error {
	server := http.Server{
		Addr:    a.config.port,
		Handler: h,
	}
	log.Printf("Server has started at address %s", a.config.port)
	return server.ListenAndServe()
}

func (a *App) mount() http.Handler {
	router := chi.NewRouter()
	
	// Initialize repo
	repo := repo.New(a.conn)
	// Initialize services 
	authSvc := auth.NewService(repo)
	userSvc := users.NewService(repo)
	formSvc := forms.NewService(repo)

	// Middlewares
	router.Use(middleware.Logger)
	router.Use(handlePreflight)
	router.Use(AuthMiddleware(authSvc))

	// Routes
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Health check!")
		w.Write([]byte("All good!"))
	})
	// Auth
	authHandler := auth.NewHandler(authSvc, userSvc, formSvc)
	router.Route("/auth", func(r chi.Router) {
		r.Get("/me", http.HandlerFunc(authHandler.Bootstrap))
	})
	// Forms
	formsHandler := forms.NewHandler(formSvc)
	router.Route("/form", func(r chi.Router) {
		r.Post("/", formsHandler.CreateForm)
		r.Put("/", formsHandler.UpdateFormSchema)
		r.Delete("/{id}", formsHandler.DeleteForm)
	})
	// User
	// userHandler := users.NewHandler(userSvc)
	// router.Route("/user", func(r chi.Router) {
	// 	r.Get("/{id}", userHandler.GetUserById)
	// 	r.Post("/create", userHandler.CreateUser)
	// 	// r.Post("/createanon", userHandler.CreateAnonUser)
	// })
	return router
}
