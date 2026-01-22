package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Migration represents a database migration file
type Migration struct {
	Name    string
	Path    string
	Content string
}

func main() {
	// Parse flags
	dsn := flag.String("dsn", "", "Database DSN (mysql://user:pass@host:port/dbname)")
	dir := flag.String("dir", "migrations", "Migrations directory")
	seed := flag.Bool("seed", false, "Run seed files after migrations")
	flag.Parse()

	// Get DSN from flag or environment
	dbDSN := *dsn
	if dbDSN == "" {
		dbDSN = os.Getenv("DATABASE_URL")
	}
	if dbDSN == "" {
		// Build from individual env vars
		host := getEnvOrDefault("DB_HOST", "localhost")
		port := getEnvOrDefault("DB_PORT", "3306")
		user := getEnvOrDefault("DB_USER", "root")
		pass := os.Getenv("DB_PASSWORD")
		name := getEnvOrDefault("DB_NAME", "cruise_price")
		dbDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", user, pass, host, port, name)
	}

	// Connect to database
	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database successfully")

	// Create migrations table if not exists
	if err := createMigrationsTable(db); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// Get executed migrations
	executed, err := getExecutedMigrations(db)
	if err != nil {
		log.Fatalf("Failed to get executed migrations: %v", err)
	}

	// Load migration files
	migrations, err := loadMigrations(*dir, false)
	if err != nil {
		log.Fatalf("Failed to load migrations: %v", err)
	}

	// Run pending migrations
	for _, m := range migrations {
		if executed[m.Name] {
			log.Printf("Skipping already executed migration: %s", m.Name)
			continue
		}

		log.Printf("Running migration: %s", m.Name)
		if err := runMigration(db, m); err != nil {
			log.Fatalf("Migration failed: %s - %v", m.Name, err)
		}
		log.Printf("Migration completed: %s", m.Name)
	}

	// Run seed files if requested
	if *seed {
		seeds, err := loadMigrations(*dir, true)
		if err != nil {
			log.Fatalf("Failed to load seeds: %v", err)
		}

		for _, s := range seeds {
			log.Printf("Running seed: %s", s.Name)
			if err := runSeed(db, s); err != nil {
				log.Fatalf("Seed failed: %s - %v", s.Name, err)
			}
			log.Printf("Seed completed: %s", s.Name)
		}
	}

	log.Println("All migrations completed successfully")
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`
	_, err := db.Exec(query)
	return err
}

func getExecutedMigrations(db *sql.DB) (map[string]bool, error) {
	executed := make(map[string]bool)

	rows, err := db.Query("SELECT name FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		executed[name] = true
	}

	return executed, rows.Err()
}

func loadMigrations(dir string, seedOnly bool) ([]Migration, error) {
	var migrations []Migration

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".sql") {
			return nil
		}

		isSeed := strings.HasPrefix(d.Name(), "seed_")
		if seedOnly != isSeed {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		migrations = append(migrations, Migration{
			Name:    d.Name(),
			Path:    path,
			Content: string(content),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by name (numeric prefix)
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name < migrations[j].Name
	})

	return migrations, nil
}

func runMigration(db *sql.DB, m Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	start := time.Now()

	// Execute migration
	if _, err = tx.Exec(m.Content); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration
	if _, err = tx.Exec("INSERT INTO schema_migrations (name) VALUES (?)", m.Name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Migration %s completed in %v", m.Name, time.Since(start))
	return nil
}

func runSeed(db *sql.DB, s Migration) error {
	start := time.Now()

	// Seeds are idempotent, no transaction tracking needed
	if _, err := db.Exec(s.Content); err != nil {
		return fmt.Errorf("failed to execute seed: %w", err)
	}

	log.Printf("Seed %s completed in %v", s.Name, time.Since(start))
	return nil
}
