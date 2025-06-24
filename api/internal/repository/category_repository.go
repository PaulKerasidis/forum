package repository

import (
	"database/sql"
	"errors"

	"github.com/PaulKerasidis/forum/internal/models"
)

type CategoryRepository struct {
	DB *sql.DB
}

// NewCategoryRepository creates a new CategoryRepository
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{DB: db}
}

// GetCategoryID validates that a category exists and returns its ID
func (cr *CategoryRepository) GetCategoryID(name string) (string, error) {
	var id string
	err := cr.DB.QueryRow("SELECT category_id FROM categories WHERE category_name = ?", name).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("category not found")
		}
		return "", err
	}

	return id, nil

}

// GetAllCategories retrieves all categories from the database
func (cr *CategoryRepository) GetAllCategories() ([]models.Category, error) {

	rows, err := cr.DB.Query("SELECT category_id, category_name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil

}

// // GetCategoryName retrieves the category name by its ID
// func (cr *CategoryRepository) GetCategoryName(id string) (string, error) {

// 	var name string
// 	err := cr.DB.QueryRow("SELECT category_name FROM categories WHERE category_id = ?", id).Scan(&name)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return "", errors.New("category not found")
// 		}
// 		return "", err
// 	}
// 	return name, nil

// }
