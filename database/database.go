package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// initDB initializes the database and returns a connection
func InitDB() (*sql.DB, error) {
	// Create database directory if it doesn't exist
	dbDir := "./database"
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}

	// Connect to SQLite database (will be created if it doesn't exist)
	dbPath := filepath.Join(dbDir, "forum.db")
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Verify connection works
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Set some basic connection pool settings
	db.SetMaxOpenConns(10)

	// Initialize database schema and data
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

	return db, nil
}

func createTables(db *sql.DB) error {
	// Start a transaction for atomicity
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Define all table creation SQL statements
	tableStatements := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS user (
			user_id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,

		// User authentication table
		`CREATE TABLE IF NOT EXISTS user_auth (
			user_id TEXT PRIMARY KEY,
			password_hash TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
		);`,

		// Sessions table
		`CREATE TABLE IF NOT EXISTS sessions (
			user_id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL UNIQUE,
			ip_address TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
		);`,

		// Categories table
		`CREATE TABLE IF NOT EXISTS categories (
			category_id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		);`,

		// Posts table
		`CREATE TABLE IF NOT EXISTS posts (
			post_id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			category_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES categories(category_id) ON DELETE CASCADE
		);`,

		// Comments table
		`CREATE TABLE IF NOT EXISTS comments (
			comment_id TEXT PRIMARY KEY,
			post_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
		);`,

		// Reactions table
		`CREATE TABLE IF NOT EXISTS reactions (
			user_id TEXT NOT NULL,
			reaction_type INTEGER NOT NULL, -- 1 for like, 2 for dislike
			comment_id TEXT,
			post_id TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, comment_id, post_id),
			FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE,
			FOREIGN KEY (comment_id) REFERENCES comments(comment_id) ON DELETE CASCADE,
			FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
			CHECK ((post_id IS NULL AND comment_id IS NOT NULL) OR (post_id IS NOT NULL AND comment_id IS NULL))
		);`,
	}

	// Execute each table creation statement
	for _, stmt := range tableStatements {
		_, err = tx.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %s: %v", stmt, err)
		}
	}

	// Commit transaction
	return tx.Commit()
}

func createIndexes(db *sql.DB) error {
	// Start a transaction for atomicity
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Define all index creation SQL statements
	indexStatements := []string{
		`CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_posts_category_id ON posts(category_id);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_reactions_user_id ON reactions(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_reactions_post_id ON reactions(post_id);`,
		`CREATE INDEX IF NOT EXISTS idx_reactions_comment_id ON reactions(comment_id);`,
	}

	// Execute each index creation statement
	for _, stmt := range indexStatements {
		_, err = tx.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %s: %v", stmt, err)
		}
	}

	// Commit transaction
	return tx.Commit()
}

func populateCategories(db *sql.DB) error {
	// Check if categories already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing categories: %v", err)
	}

	// If categories already exist, skip population
	if count > 0 {
		fmt.Println("Categories already exist, skipping population.")
		return nil
	}

	// Define the categories to add
	categories := []string{
		"General Discussion",
		"Programming",
		"Golang",
		"Web Development",
		"Database Systems",
		"DevOps",
		"Mobile Development",
		"Machine Learning",
		"Security",
		"Off-Topic",
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Prepare the insert statement
	stmt, err := tx.Prepare("INSERT INTO categories (name) VALUES (?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	// Insert each category
	for _, category := range categories {
		_, err = stmt.Exec(category)
		if err != nil {
			return fmt.Errorf("failed to insert category '%s': %v", category, err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
