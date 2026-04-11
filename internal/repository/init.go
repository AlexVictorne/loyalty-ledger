package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositorySet struct {
	User UserRepository
	PGX  *pgxpool.Pool // nil если in-memory
}

// Инициализируем репозитории и, если указана работа через БД, подключение к БД и миграции.
func InitRepositories(databaseURI string, migrateFunc func(db *sql.DB) error) (*RepositorySet, error) {
	if databaseURI != "" {
		db, err := sql.Open("postgres", databaseURI)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to db: %w", err)
		}
		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("db not available: %w", err)
		}
		if err := ApplyMigrations(db, "migrations"); err != nil {
			return nil, fmt.Errorf("migration failed: %w", err)
		}
		db.Close()
		pool, err := pgxpool.New(context.Background(), databaseURI)
		if err != nil {
			return nil, fmt.Errorf("failed to create pgx pool: %w", err)
		}
		if err := pool.Ping(context.Background()); err != nil {
			pool.Close()
			return nil, fmt.Errorf("pgx pool not available: %w", err)
		}
		return &RepositorySet{
			User: NewPostgresUserRepository(pool),
			PGX:  pool,
		}, nil
	}
	return &RepositorySet{
		User: NewInMemoryUserRepository(),
		PGX:  nil,
	}, nil
}

func ApplyMigrations(db *sql.DB, migrationsDir string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsDir,
		"postgres", driver,
	)
	if err != nil {
		return err
	}

	err = m.Up()
	if err == nil {
		log.Println("[migrate] DB up successfully")
		return nil
	}
	if errors.Is(err, migrate.ErrNoChange) {
		ver, dirty, verr := m.Version()
		if verr != nil {
			log.Printf("[migrate] DB actual, but version incorrect %v", verr)
		} else {
			log.Printf("[migrate] DB actual, current version: %d (dirty=%v)", ver, dirty)
		}
		return nil
	}
	return err
}
