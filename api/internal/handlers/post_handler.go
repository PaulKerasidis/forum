package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/PaulKerasidis/forum/config"
	"github.com/PaulKerasidis/forum/internal/middleware"
	"github.com/PaulKerasidis/forum/internal/models"
	"github.com/PaulKerasidis/forum/internal/repository"
	"github.com/PaulKerasidis/forum/internal/utils"
)

// ...
// CRUD HANDLERS FOR POSTS
// ...
// REPLACE the CreatePostHandler in post_handler.go:

func CreatePostHandler(pr *repository.PostsRepository, cr *repository.CategoryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get authenticated user
		user := middleware.GetCurrentUser(r)
		if user == nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		var req models.CreatePostRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		if utils.ValidatePostContent(req.Content) != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid post content")
			return
		}

		// Validate and get category IDs
		var categoryIDs []string
		for _, categoryName := range req.CategoryNames {
			categoryID, err := cr.GetCategoryID(categoryName)
			if err != nil {
				utils.RespondWithError(w, http.StatusBadRequest, "Invalid category: "+categoryName)
				return
			}
			categoryIDs = append(categoryIDs, categoryID)
		}
		if len(categoryIDs) < config.Config.MinCategories {
			utils.RespondWithError(w, http.StatusBadRequest,
				fmt.Sprintf("Minimum %d category required", config.Config.MinCategories))
			return
		}

		if len(categoryIDs) > config.Config.MaxCategories {
			utils.RespondWithError(w, http.StatusBadRequest,
				fmt.Sprintf("Maximum %d categories allowed", config.Config.MaxCategories))
			return
		}
		// Create post - now returns lightweight response
		createResponse, err := pr.CreatePost(user.ID, req.Content, categoryIDs)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create post")
			return
		}

		// Return lightweight response
		utils.RespondWithSuccess(w, http.StatusCreated, createResponse)
	}
}

// UpdatePostHandler updates an existing post
func UpdatePostHandler(pr *repository.PostsRepository, cr *repository.CategoryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		// Get authenticated user
		user := middleware.GetCurrentUser(r)
		if user == nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
		// Extract post ID from URL path
		postID := r.PathValue("id")
		if postID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Post ID is required")
			return
		}

		// Pass userID to GetPostByID for ownership check
		post, err := pr.GetPostByID(postID, &user.ID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		if post.UserID != user.ID {
			utils.RespondWithError(w, http.StatusForbidden, "You can only update your own posts")
			return
		}
		// Parse request body
		var req models.UpdatePostRequest
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		// Validate and get category IDs
		var categoryIDs []string
		for _, categoryName := range req.CategoryNames {
			categoryID, err := cr.GetCategoryID(categoryName)
			if err != nil {
				utils.RespondWithError(w, http.StatusBadRequest, "Invalid category: "+categoryName)
				return
			}
			categoryIDs = append(categoryIDs, categoryID)
		}
		if len(categoryIDs) < config.Config.MinCategories {
			utils.RespondWithError(w, http.StatusBadRequest,
				fmt.Sprintf("Minimum %d category required", config.Config.MinCategories))
			return
		}

		if len(categoryIDs) > config.Config.MaxCategories {
			utils.RespondWithError(w, http.StatusBadRequest,
				fmt.Sprintf("Maximum %d categories allowed", config.Config.MaxCategories))
			return
		}
		if utils.ValidatePostContent(req.Content) != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid post content")
			return
		}
		// Update post
		err = pr.UpdatePost(postID, user.ID, req.Content, categoryIDs)
		if err != nil {
			if err.Error() == "post not found" {
				utils.RespondWithError(w, http.StatusNotFound, "Post not found")
				return
			}
			if err.Error() == "unauthorized: you can only update your own posts" {
				utils.RespondWithError(w, http.StatusForbidden, "You can only update your own posts")
				return
			}
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update post")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, nil)
	}
}

// DeletePostHandler deletes an existing post
func DeletePostHandler(pr *repository.PostsRepository, cr *repository.CategoryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		// Get authenticated user
		user := middleware.GetCurrentUser(r)
		if user == nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
		// Extract post ID from URL path
		postID := r.PathValue("id")
		if postID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Post ID is required")
			return
		}

		// Pass userID to GetPostByID for ownership check
		post, err := pr.GetPostByID(postID, &user.ID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		// Check if the post belongs to the user
		if post.UserID != user.ID {
			utils.RespondWithError(w, http.StatusForbidden, "You can only delete your own posts")
			return
		}
		// Delete the post
		err = pr.DeletePost(postID, user.ID)
		if err != nil {
			if err.Error() == "post not found" {
				utils.RespondWithError(w, http.StatusNotFound, "Post not found")
				return
			}
			if err.Error() == "unauthorized: you can only delete your own posts" {
				utils.RespondWithError(w, http.StatusForbidden, "You can only delete your own posts")
				return
			}
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete post")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, nil)
	}
}

// ...
// GET HANDLERS FOR POSTS - UPDATED WITH USER CONTEXT
// ...
// GetAllPostsHandler retrieves all posts with pagination and sorting
func GetAllPostsHandler(pr *repository.PostsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get authenticated user
		currentUser := middleware.GetCurrentUser(r)
		var userID *string = nil
		if currentUser != nil {
			userID = &currentUser.ID
		}

		// Parse pagination parameters - ONE LINE!
		limit, offset := utils.ParsePaginationParams(r)
		// Parse sort options from query parameters
		sortOptions := utils.ParsePostSortOptions(r)
		// Get posts and total count
		posts, err := pr.GetAllPosts(limit, offset, userID, sortOptions)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve posts")
			return
		}

		totalCount, err := pr.GetCountTotalPosts()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve posts count")
			return
		}

		// Respond with standardized format - ONE LINE!
		utils.RespondWithPaginatedPosts(w, posts, totalCount, limit, offset)
	}
}

// GetPostsByCategoryHandler retrieves posts filtered by category with pagination and sorting
func GetPostsByCategoryHandler(pr *repository.PostsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Check for authenticated user
		currentUser := middleware.GetCurrentUser(r)
		var userID *string = nil
		if currentUser != nil {
			userID = &currentUser.ID
		}

		// Extract category ID from URL path
		categoryID := r.PathValue("id")
		if categoryID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Category ID is required")
			return
		}
		// Parse pagination parameters
		limit, offset := utils.ParsePaginationParams(r)
		// Parse sort options from query parameters
		sortOptions := utils.ParsePostSortOptions(r)
		// Get total count for this category first
		totalCount, err := pr.GetCountPostByCategory(categoryID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve posts count")
			return
		}

		// Pass userID to repository
		posts, err := pr.GetPostsByCategory(categoryID, limit, offset, userID, sortOptions)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve posts")
			return
		}
		utils.RespondWithPaginatedPosts(w, posts, totalCount, limit, offset)
	}
}

// GetSinglePostHandler retrieves a single post with full details and reactions
func GetSinglePostHandler(pr *repository.PostsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		//  Check for authenticated user
		currentUser := middleware.GetCurrentUser(r)
		var userID *string = nil
		if currentUser != nil {
			userID = &currentUser.ID
		}

		// Extract post ID from URL path
		postID := r.PathValue("id")
		if postID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Post ID is required")
			return
		}

		//Pass userID to repository
		post, err := pr.GetPostByID(postID, userID)
		if err != nil {
			if err.Error() == "post not found" {
				utils.RespondWithError(w, http.StatusNotFound, "Post not found")
				return
			}
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve post")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, post)
	}
}

// ...
// PROFILE HANDLERS - UPDATED
// ...
// GetUserPostsProfileHandler retrieves all posts by a specific user
func GetUserPostsProfileHandler(pr *repository.PostsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		// Get authenticated user
		currentUser := middleware.GetCurrentUser(r)
		if currentUser == nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
		// Extract user ID from URL path
		userID := r.PathValue("id")
		if userID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
			return
		}
		// Ensure user can only view their own posts
		if currentUser.ID != userID {
			utils.RespondWithError(w, http.StatusForbidden, "You can only view your own posts")
			return
		}
		totalCount, err := pr.GetCountPostByUser(userID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve posts count")
			return
		}
		limit, offset := utils.ParsePaginationParams(r)
		sortOptions := utils.ParsePostSortOptions(r)
		//Pass both targetUserID and currentUserID to repository
		posts, err := pr.GetPostsByUser(userID, limit, offset, &currentUser.ID, sortOptions)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user posts")
			return
		}
		utils.RespondWithPaginatedPosts(w, posts, totalCount, limit, offset)
	}
}

// GetUserLikedPostsProfileHandler retrieves all posts liked by a specific user
func GetUserLikedPostsProfileHandler(pr *repository.PostsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		// Get authenticated user
		currentUser := middleware.GetCurrentUser(r)
		if currentUser == nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
		// Extract user ID from URL path
		userID := r.PathValue("id")
		if userID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
			return
		}
		// Ensure user can only view their own liked posts
		if currentUser.ID != userID {
			utils.RespondWithError(w, http.StatusForbidden, "You can only view your own liked posts")
			return
		}
		// Parse pagination parameters
		limit, offset := utils.ParsePaginationParams(r)
		// sorting options
		sortOptions := utils.ParsePostSortOptions(r)
		totalCount, err := pr.GetCountLikedPostByUser(userID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve liked posts count")
			return
		}
		// Pass both targetUserID and currentUserID to repository
		posts, err := pr.GetPostsLikedByUser(userID, limit, offset, &currentUser.ID, sortOptions)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve liked posts")
			return
		}
		utils.RespondWithPaginatedPosts(w, posts, totalCount, limit, offset)
	}
}

// GetUserCommentedPostsProfileHandler retrieves all posts that a specific user has commented on
func GetUserCommentedPostsProfileHandler(pr *repository.PostsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get authenticated user
		currentUser := middleware.GetCurrentUser(r)
		if currentUser == nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Extract user ID from URL path
		userID := r.PathValue("id")
		if userID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
			return
		}

		// Ensure user can only view their own commented posts
		if currentUser.ID != userID {
			utils.RespondWithError(w, http.StatusForbidden, "You can only view your own commented posts")
			return
		}

		// Parse pagination parameters
		limit, offset := utils.ParsePaginationParams(r)
		// Parse sort options from query parameters
		sortOptions := utils.ParsePostSortOptions(r)
		// Get total count of posts user has commented on
		totalCount, err := pr.GetCountCommentedPostByUser(userID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve commented posts count")
			return
		}

		// Get posts that user has commented on
		posts, err := pr.GetPostsCommentedByUser(userID, limit, offset, &currentUser.ID, sortOptions)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve commented posts")
			return
		}

		// Respond with paginated posts
		utils.RespondWithPaginatedPosts(w, posts, totalCount, limit, offset)
	}
}
