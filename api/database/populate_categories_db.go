package database

import (
	"database/sql"
	"fmt"

	"github.com/PaulKerasidis/forum/internal/utils"
)

func populateCategories(db *sql.DB) error {
	// Check if categories already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing categories: %v", err)
	}

	// If categories already exist, skip population
	if count > 0 {
		return nil
	}

	// Define the categories to add
	categories := []string{
		"General discussion",
		"Programming",
		"Web Development",
		"Gardening",
		"Networking",
		"Fishing",
		"Game Development",
		"Database Management",
		"DevOps",
		"Cloud Computing",
		"Mobile Development",
		"Machine Learning",
		"Security",
		"Off-Topic",
	}
	// Ensure texists

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Prepare the insert statement - now category_id is auto-incremented
	stmt, err := tx.Prepare("INSERT INTO categories (category_id, category_name) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()
	// Insert each category
	for _, category := range categories {
		category_id := utils.GenerateUUIDToken() // Generate a unique ID for the category
		_, err = stmt.Exec(category_id, category)
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
