package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"scratch/api"
	"scratch/internal"
	"scratch/internal/authorization/session"
	"scratch/internal/services"
	storage "scratch/internal/storage/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

//go:embed internal/storage/migrations/*
var embedMigrations embed.FS

func setupAppHandler() http.Handler {
	r := chi.NewRouter()

	database, err := initDatabase()

	defer database.Close()

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		os.Exit(1)
	}

	s := session.NewJsonWebToken(session.Config{TokenSecret: []byte(os.Getenv("JWT_SECRET"))})

	accountService := services.NewAccountService(storage.New(database), s, slog.Logger{})

	ah := internal.NewAccountHandler(accountService, slog.Logger{})

	server := api.HandlerWithOptions(ah, api.ChiServerOptions{
		BaseRouter: r,
		Middlewares: []api.MiddlewareFunc{
			middleware.Logger,
		},
	})

	return server
}
func main() {
	h := setupAppHandler()
	log.Fatalln(http.ListenAndServe("localhost:8080", h))
}

func initDatabase() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%v port=%v user=%v "+
		"password=%v dbname=%v sslmode=disable",
		os.Getenv("HOST"), os.Getenv("PORT"),
		os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("DBNAME"))

	database, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("open connection to database: %w", err)
	}

	err = database.Ping()
	if err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	err = runUpMigrations(database)
	if err != nil {
		return nil, fmt.Errorf("setup migrations: %w", err)
	}

	return database, nil
}
func runUpMigrations(database *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose - set dialect: %w", err)
	}

	if err := goose.Up(database, "internal/storage/migrations"); err != nil {
		return fmt.Errorf("run up migrations: %w", err)
	}
	return nil
}

func runDownMigrations(database *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose - set dialect: %w", err)
	}

	if err := goose.DownTo(database, "internal/storage/migrations", 20231006202055); err != nil {
		return fmt.Errorf("run up migrations: %w", err)
	}
	return nil
}

/*
sklep internetowy
log z handlerem ktory wrzuca do pliku - slog.NewJSON...
- code coverage
- dodanie usera przez CMD
- logowanie, rejestracja, sesja uzytkownika
- paseto, session
- formularz w htmx
- logger
- jager
- emaile z asynq redis
- wrzucanie na kafke
- handlowanie secretow
- k8s
- docker
- sentry
- lintery
*/
