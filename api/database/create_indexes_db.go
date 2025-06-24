package database

import (
	"database/sql"
	"fmt"
)

func createIndexes(db *sql.DB) error {
	// Start a transaction for atomicity
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Execute each index creation statement from the IndexCreationStatements slice
	for _, stmt := range IndexCreationStatements {
		_, err = tx.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %s: %v", stmt, err)
		}
	}

	// Commit transaction
	return tx.Commit()
}
