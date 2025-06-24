package handlers

import (
	"log"
	"net/http"

	"frontend-service/internal/models"
	"frontend-service/internal/services"
	"frontend-service/internal/session"
	"frontend-service/internal/utils"
)

type CategoryHandler struct {
	authService     *services.AuthService
	postService     *services.PostService
	categoryService *services.CategoryService
	templateService *services.TemplateService
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(authService *services.AuthService, postService *services.PostService, categoryService *services.CategoryService, templateService *services.TemplateService) *CategoryHandler {
	return &CategoryHandler{
		authService:     authService,
		postService:     postService,
		categoryService: categoryService,
		templateService: templateService,
	}
}

// Update your frontend category_handler.go:

func (h *CategoryHandler) ServeCategoryPosts(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user
	user := session.GetUserFromSession(r, h.authService)

	// Extract category ID from URL
	categoryID := r.PathValue("id")
	if categoryID == "" {
		http.Error(w, "Category ID required", http.StatusBadRequest)
		return
	}

	// Parse pagination and sort parameters
	pagination := utils.ParsePaginationFromRequest(r)
	sortBy := utils.ParseSortFromRequest(r, "newest")

	// Get session cookie for API requests
	sessionCookie, _ := session.GetSessionCookie(r, h.authService)

	// Get posts by category from backend
	categoryData, err := h.postService.GetPostsByCategory(categoryID, pagination.Limit, pagination.Offset, sortBy, sessionCookie)
	if err != nil {
		log.Printf("Error fetching category posts: %v", err)
		http.Error(w, "Failed to load posts", http.StatusInternalServerError)
		return
	}

	// Get all categories for sidebar
	categories, err := h.categoryService.GetCategories()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		categories = []models.Category{} // Empty fallback
	}

	// 🎯 SMART SOLUTION: Extract category info from the posts data
	var currentCategory *models.Category
	if len(categoryData.Posts) > 0 && len(categoryData.Posts[0].Categories) > 0 {
		// Get category info from the first post
		firstPostCategory := categoryData.Posts[0].Categories[0]
		currentCategory = &models.Category{
			ID:    firstPostCategory.ID,
			Name:  firstPostCategory.Name,
			Count: categoryData.Pagination.TotalCount, // ✅ Use total count from pagination!
		}
	} else {
		// Fallback: If no posts, find category by ID from all categories
		for _, cat := range categories {
			if cat.ID == categoryID {
				currentCategory = &models.Category{
					ID:    cat.ID,
					Name:  cat.Name,
					Count: 0, // No posts in this category
				}
				break
			}
		}
	}

	// Handle case where category not found
	if currentCategory == nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	// 🔧 FIX: Convert CategoryPostsResponse to PaginatedPostsResponse
	postsData := &models.PaginatedPostsResponse{
		Posts:      categoryData.Posts,      // Extract posts array
		Pagination: categoryData.Pagination, // Extract pagination info
	}

	// Prepare template data
	data := models.CategoryPageData{
		Category:   currentCategory,
		Posts:      postsData,
		Categories: categories,
		User:       user,
	}

	// Render template
	if err := h.templateService.Render(w, "category.html", data); err != nil {
		log.Printf("Error rendering category template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}
