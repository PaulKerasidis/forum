package utils

import "database/sql"

// ExecuteInTransaction runs a function inside a transaction
// Use this for operations that only return an error (like delete, update without return value)
func ExecuteInTransaction(db *sql.DB, fn func(*sql.Tx) error) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Ensure rollback is called if something goes wrong
	defer tx.Rollback()

	// Execute the function
	if err := fn(tx); err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

// ExecuteInTransactionWithResult runs a function inside a transaction and returns a result
// Use this for operations that return something (like create, get operations)
// The [T any] makes it work with any return type: *models.User, []models.Category, string, bool, etc.
func ExecuteInTransactionWithResult[T any](db *sql.DB, fn func(*sql.Tx) (T, error)) (T, error) {
	var zero T // zero value for type T (nil for pointers, empty slice, etc.)

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return zero, err
	}

	// Ensure rollback is called if something goes wrong
	defer tx.Rollback()
	
	// Execute the function and get result
	result, err := fn(tx)
	if err != nil {
		return zero, err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return zero, err
	}

	return result, nil
}
