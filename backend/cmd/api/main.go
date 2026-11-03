package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"finance-backend/internal/api"
	"finance-backend/internal/store"
	"finance-backend/migrations"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	if err := runMigrationsWithRetry(dbURL); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("migrations applied")

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer pool.Close()

	srv := api.NewServer(store.New(pool))

	addr := ":" + port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		log.Fatal(err)
	}
}

// runMigrationsWithRetry retries the initial connection: right after the
// Postgres container's healthcheck passes, container DNS can briefly still
// fail to resolve the db hostname (observed in Compose), so depends_on
// alone isn't enough.
func runMigrationsWithRetry(dbURL string) error {
	const attempts = 10
	var err error
	for i := 1; i <= attempts; i++ {
		if err = runMigrations(dbURL); err == nil {
			return nil
		}
		if i < attempts {
			time.Sleep(time.Second)
		}
	}
	return err
}

// runMigrations applies pending migrations embedded in the binary, using
// golang-migrate as a library so no separate `migrate` CLI step is needed
// either natively or in a container.
func runMigrations(dbURL string) error {
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("source driver: %w", err)
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
