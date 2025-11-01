package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/config"
	_ "modernc.org/sqlite"
)

// New creates a new database connection
func New(cfg config.DatabaseConfig) (*sql.DB, error) {
	var db *sql.DB
	var err error

	if cfg.Driver == "sqlite" {
		db, err = sql.Open("sqlite", cfg.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite database: %w", err)
		}

		// Enable WAL mode for better concurrency
		if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
			return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
		}

		// Set connection pool settings
		db.SetMaxOpenConns(1) // SQLite only supports one writer at a time
		db.SetMaxIdleConns(1)
		db.SetConnMaxLifetime(time.Hour)

	} else if cfg.Driver == "postgres" {
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
		)

		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open postgres database: %w", err)
		}

		// Set connection pool settings
		db.SetMaxOpenConns(cfg.MaxOpenConns)
		db.SetMaxIdleConns(cfg.MaxIdleConns)
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	} else {
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
