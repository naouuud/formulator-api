package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	godotenv.Load()
	config := AppConfig{
		port: ":8080",
		dsn:  os.Getenv("GOOSE_DBSTRING"),
	}
	slogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(slogger)
	conn, err := pgx.Connect(ctx, config.dsn)
	if err != nil {
		slog.Error("Error connecting to database", "error", err)
		panic(err)
	}
	defer log.Println("Database connection closed")
	defer conn.Close(ctx)
	log.Println("Connected to database")
	app := App{
		config: config,
		conn:   conn,
	}
	// run application
	if err := app.run(app.mount()); err != nil {
		// log.Fatalf("Error starting server: %s", err)
		slog.Error("Error starting server", "error", err)
		os.Exit(1)
	}

	// Mini app
	// router := chi.NewRouter()
	// router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("Simple health check, all good!"))
	// })
	// _, err := pgx.Connect(context.Background(), os.Getenv("GOOSE_DBSTRING"))
	// if err != nil {
	// 	log.Fatal("Error connecting to db")
	// }
	// http.ListenAndServe(":8080", router)
}
