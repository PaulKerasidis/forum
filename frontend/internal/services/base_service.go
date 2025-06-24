package services

import (
	"net/http"
	"time"
)

// BaseClient provides shared HTTP client functionality
type BaseClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewBaseClient creates a new base client
func NewBaseClient(baseURL string) *BaseClient {
	return &BaseClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}
