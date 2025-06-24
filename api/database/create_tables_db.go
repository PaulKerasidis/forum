package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func createTables(db *sql.DB) error {
	// Start a transaction for atomicity
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Execute each table creation statement from the TableCreationStatements slice
	for _, stmt := range TableCreationStatements {
		_, err = tx.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %s: %v", stmt, err)
		}
	}

	// Commit transaction
	return tx.Commit()
}
