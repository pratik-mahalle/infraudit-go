package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/pratik-mahalle/infraudit/internal/config"
	"github.com/pratik-mahalle/infraudit/internal/repository/postgres"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Connect to database
	db, err := postgres.New(cfg.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Println("Connected to database successfully")

	// Create migrations table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255) NOT NULL UNIQUE,
			executed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create migrations table: %v\n", err)
		os.Exit(1)
	}

	// Get migration files
	migrationsDir := "./migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read migrations directory: %v\n", err)
		os.Exit(1)
	}

	// Sort files to ensure correct order
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".sql" {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	if len(sqlFiles) == 0 {
		fmt.Println("No migration files found")
		return
	}

	// Run migrations
	for _, filename := range sqlFiles {
		// Check if migration already executed
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = ?", filename).Scan(&count)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check migration status: %v\n", err)
			os.Exit(1)
		}

		if count > 0 {
			fmt.Printf("Skipping %s (already executed)\n", filename)
			continue
		}

		// Read migration file
		content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read migration file %s: %v\n", filename, err)
			os.Exit(1)
		}

		// Execute migration
		fmt.Printf("Running migration: %s\n", filename)
		_, err = db.Exec(string(content))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to execute migration %s: %v\n", filename, err)
			os.Exit(1)
		}

		// Record migration
		_, err = db.Exec("INSERT INTO migrations (name) VALUES (?)", filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to record migration %s: %v\n", filename, err)
			os.Exit(1)
		}

		fmt.Printf("✓ Migration %s completed successfully\n", filename)
	}

	fmt.Println("\nAll migrations completed successfully!")
}

func getMigratedVersions(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT name FROM migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	migrated := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		migrated[name] = true
	}

	return migrated, rows.Err()
}
