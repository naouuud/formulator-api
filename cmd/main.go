package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	config := AppConfig{
		port: ":8080",
		dsn: "host=localhost port=5434 user=naoude password=6%71hiu& dbname=formulator sslmode=disable",
	}
	// slogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// slog.SetDefault(slogger)
	conn, err := pgx.Connect(ctx, config.dsn)
	if err != nil {
		// slogger.Error("Unable to connect to db", "error", err)
		log.Panicf("Error connecting to database: %s", err)
	}
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
	// _, err := pgx.Connect(context.Background(), "host=localhost port=5434 user=naoude password=6%71hiu& dbname=formulator sslmode=disable")
	// if err != nil {
	// 	log.Fatal("Error connecting to db")
	// }
	// http.ListenAndServe(":8080", router)
}
