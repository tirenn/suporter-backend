package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"suporter-backend/internal/config"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	if err := ensureDatabaseExists(cfg); err != nil {
		log.Printf("[DB Warning] Could not ensure database exists: %v", err)
	}

	dsn := cfg.DSN()
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgresql via GORM: %w", err)
	}

	stdDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB instance: %w", err)
	}

	if err := stdDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("[GORM] Successfully connected to PostgreSQL (%s:%s/%s)", cfg.DBHost, cfg.DBPort, cfg.DBName)

	// Run Goose migrations
	if err := RunGooseMigrations(stdDB, "migrations"); err != nil {
		log.Printf("[Goose Migration Warning] %v", err)
	}

	return gormDB, nil
}

func DropAllTables(db *sql.DB) error {
	log.Println("[DB Reset] Dropping all database tables (goose_db_version, templates, alerts, projects, users)...")
	_, err := db.Exec("DROP TABLE IF EXISTS goose_db_version, templates, alerts, projects, users CASCADE;")
	if err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}
	log.Println("[DB Reset] Database tables dropped cleanly.")
	return nil
}

func ensureDatabaseExists(cfg *config.Config) error {
	rootDB, err := sql.Open("postgres", cfg.RootDSN())
	if err != nil {
		return err
	}
	defer rootDB.Close()

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1);`
	if err := rootDB.QueryRow(query, cfg.DBName).Scan(&exists); err != nil {
		return err
	}

	if !exists {
		log.Printf("[DB] Database '%s' does not exist. Creating...", cfg.DBName)
		_, err := rootDB.Exec(fmt.Sprintf("CREATE DATABASE %s;", cfg.DBName))
		if err != nil {
			return fmt.Errorf("failed to create database '%s': %w", cfg.DBName, err)
		}
		log.Printf("[DB] Database '%s' created successfully.", cfg.DBName)
	}
	return nil
}

func RunGooseMigrations(db *sql.DB, migrationsDir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose dialect error: %w", err)
	}

	dir := migrationsDir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fallback := filepath.Join("..", "..", migrationsDir)
		if _, err2 := os.Stat(fallback); err2 == nil {
			dir = fallback
		}
	}

	log.Printf("[Goose] Running database migrations from '%s'...", dir)
	if err := goose.Up(db, dir); err != nil {
		log.Println("[Goose Reset] Schema mismatch detected. Resetting database tables for new migration sequence...")
		_ = DropAllTables(db)
		if err2 := goose.Up(db, dir); err2 == nil {
			log.Println("[Goose] Clean schema rebuild succeeded.")
			return nil
		}
		return fmt.Errorf("goose up migration failed: %w", err)
	}

	log.Println("[Goose] All migrations applied successfully.")
	return nil
}

func RunGooseDown(db *sql.DB, migrationsDir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Down(db, migrationsDir)
}

func CreateGooseMigration(migrationsDir string, name string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	dir := migrationsDir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fallback := filepath.Join("..", "..", migrationsDir)
		if _, err2 := os.Stat(fallback); err2 == nil {
			dir = fallback
		}
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("migration name cannot be empty (usage: make migrate-create name=migration_name)")
	}

	log.Printf("[Goose] Creating SQL migration file '%s' in '%s'...", name, dir)
	return goose.Create(nil, dir, name, "sql")
}
