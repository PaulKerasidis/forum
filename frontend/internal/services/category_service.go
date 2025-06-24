package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"frontend-service/internal/models"
)

type CategoryService struct {
	*BaseClient
}

// NewCategoryService creates a new category service
func NewCategoryService(baseClient *BaseClient) *CategoryService {
	return &CategoryService{
		BaseClient: baseClient,
	}
}

// GetCategories retrieves categories from the backend API
func (s *CategoryService) GetCategories() ([]models.Category, error) {
	// Make HTTP request
	resp, err := s.HTTPClient.Get(s.BaseURL + "/categories")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check API success
	if !apiResponse.Success {
		return nil, fmt.Errorf("API error: %s", apiResponse.Error)
	}

	// Convert data to Categories slice
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var categories []models.Category
	if err := json.Unmarshal(dataBytes, &categories); err != nil {
		return nil, fmt.Errorf("failed to parse categories data: %w", err)
	}

	return categories, nil
}
