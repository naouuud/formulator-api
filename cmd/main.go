package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	godotenv.Load()
	config := AppConfig{
		port: ":8080",
		dsn: os.Getenv("GOOSE_DBSTRING"),
	}
	// slogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// slog.SetDefault(slogger)
	conn, err := pgx.Connect(ctx, config.dsn)
	log.Println("Connected to database")
	if err != nil {
		// slogger.Error("Unable to connect to db", "error", err)
		log.Panicf("Error connecting to database: %s", err)
	}
	defer log.Println("Database connection closed")
	defer conn.Close(ctx) 
	app := App{
		config: config,
		conn: conn,
	}
	// run application
	if err := app.run(app.mount()); err != nil {
		log.Fatalf("Error starting server: %s", err)
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
