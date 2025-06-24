package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/PaulKerasidis/forum/config"
)

// InitDB initializes the database only if it does not exist
func InitDB() (*sql.DB, error) {
	dbDir := filepath.Dir(config.Config.DBPath)

	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}

	// Check if the database file exists
	dbExists := true
	if _, err := os.Stat(config.Config.DBPath); os.IsNotExist(err) {
		dbExists = false
	}

	// Connect to SQLite database
	db, err := sql.Open("sqlite3", config.Config.DBPath+"?_foreign_keys=on&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// For SQLite: Use config values for database connections
	db.SetMaxOpenConns(config.Config.DBMaxConnections)
	db.SetMaxIdleConns(config.Config.DBMaxConnections / 2) // Half of max connections for idle
	db.SetConnMaxLifetime(30 * time.Minute)                // Connections expire after 30 minutes
	db.SetConnMaxIdleTime(5 * time.Minute)                 // Idle connections timeout after 5 minutes
	// If the database didn't exist before, create schema and populate data
	if !dbExists {
		fmt.Println("Initializing new database...")

		if err := createTables(db); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create tables: %v", err)
		}

		if err := createIndexes(db); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create indexes: %v", err)
		}

		if err := populateCategories(db); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to populate categories: %v", err)
		}

		fmt.Println("Database initialized successfully.")
	} else {
		fmt.Println("Database already exists. Skipping initialization.")
	}

	return db, nil
}
