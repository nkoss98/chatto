package internal

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"scratch/api"
	"scratch/internal/authorization/session"
	"scratch/internal/services"
	storage "scratch/internal/storage/database"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/ory/dockertest/v3"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
)

var (
	localhost = "localhost"
	port      = 5432
	user      = "postgres"
	password  = "postgres"
	dbname    = "tests"
	db        *sql.DB
)

//go:embed storage/migrations/*
var embedMigrations embed.FS

func connectPostgres(pool *dockertest.Pool) (*dockertest.Resource, error) {
	resource, err := pool.Run("postgres", "13.8",
		[]string{"POSTGRES_DB=tests", "POSTGRES_PASSWORD=postgres"})
	if err != nil {
		return nil, fmt.Errorf("run new pool: %w", err)
	}

	psqlInfo := fmt.Sprintf("host=%v port=%v user=%v "+
		"password=%v dbname=%v sslmode=disable default_query_exec_mode=describe_exec",
		localhost, resource.GetPort("5432/tcp"), user, password, dbname)

	if err = pool.Retry(func() error {
		db, err = sql.Open("pgx", psqlInfo)
		if err != nil {
			return fmt.Errorf("fail to retry: %w", err)
		}
		return db.Ping()
	}); err != nil {
		return nil, fmt.Errorf("retry connection: %w", err)
	}
	return resource, nil
}

func TestMain(m *testing.M) {

	var pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalln(fmt.Sprintf("create new pool: %v", err))
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalln(fmt.Sprintf("ping pool: %v", err))
	}

	resource, err := connectPostgres(pool)
	if err != nil {
		log.Fatal(fmt.Sprintf("connect postgres: %v", err))
	}
	code := m.Run()

	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	r := chi.NewRouter()

	err := godotenv.Load(".env.test")
	if err != nil {
		assert.NoError(t, err)
	}

	s := session.NewJsonWebToken(session.Config{TokenSecret: []byte(os.Getenv("JWT_SECRET"))})

	accountService := services.NewAccountService(storage.New(db), s, slog.Logger{})

	ah := NewAccountHandler(accountService, slog.Logger{})

	server := api.HandlerWithOptions(ah, api.ChiServerOptions{
		BaseRouter: r,
		Middlewares: []api.MiddlewareFunc{
			middleware.Logger,
		},
	})

	return server

}

func initService(t *testing.T) *httptest.Server {
	err := runUpMigrations(db)
	assert.NoError(t, err)

	t.Cleanup(func() {
		err = runDownMigrations(db)
		assert.NoError(t, err)
	})

	srv := httptest.NewServer(testHandler(t))

	t.Cleanup(func() {
		srv.Close()
	})

	return srv
}

func runUpMigrations(database *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose - set dialect: %w", err)
	}

	if err := goose.Up(database, "storage/migrations"); err != nil {
		return fmt.Errorf("run up migrations: %w", err)
	}
	return nil
}

func runDownMigrations(database *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose - set dialect: %w", err)
	}

	if err := goose.DownTo(database, "storage/migrations", 0); err != nil {
		return fmt.Errorf("run up migrations: %w", err)
	}
	return nil
}
