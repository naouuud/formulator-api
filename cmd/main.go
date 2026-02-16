package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
)

type App struct {
	config    AppConfig
	conn *pgx.Conn
}

type AppConfig struct {
	port  string
	dbString string
}

func main() {
	ctx := context.Background()
	config := AppConfig{
		port: ":8080",
		dbString: "host=localhost port=5434 user=naoude password=6%71hiu& dbname=formulator sslmode=disable",
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	conn, err := pgx.Connect(ctx, config.dbString)
	if err != nil {
		logger.Error("Unable to connect to db", err)
		panic(err)
	}
	defer conn.Close(ctx)
	app := App{
		config: config,
		conn: conn,
	}
	// run application
}
