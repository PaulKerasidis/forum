package routes

import (
	"net/http"

	"frontend-service/config"
	"frontend-service/internal/handlers"
	"frontend-service/internal/services"
)

// SetupRoutes configures all routes for the frontend service
func SetupRoutes(authService *services.AuthService, postService *services.PostService, categoryService *services.CategoryService, userService *services.UserService, commentService *services.CommentService, postReactionService *services.PostReactionService, commentReactionService *services.CommentReactionService, templateService *services.TemplateService, cfg *config.Config) *http.ServeMux { // CHANGED: Added cfg parameter
	mux := http.NewServeMux()

	// Serve static files (CSS, JS, images, etc.)
	staticDir := "./web/static/"
	fs := http.FileServer(http.Dir(staticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Create handlers with config access where needed
	homeHandler := handlers.NewHomeHandler(authService, postService, categoryService, templateService)
	authHandler := handlers.NewAuthHandler(authService, templateService, cfg) // CHANGED: Pass config to auth handler
	oauthHandler := handlers.NewOAuthHandler(authService, templateService, cfg.APIBaseURL) // NEW: OAuth handler
	categoryHandler := handlers.NewCategoryHandler(authService, postService, categoryService, templateService)
	postHandler := handlers.NewPostHandler(authService, postService, templateService)
	createPostHandler := handlers.NewCreatePostHandler(authService, postService, categoryService, templateService)
	editPostHandler := handlers.NewEditPostHandler(authService, postService, categoryService, templateService)
	deletePostHandler := handlers.NewDeletePostHandler(authService, postService)
	profileHandler := handlers.NewProfileHandler(authService, userService, templateService)
	commentHandler := handlers.NewCommentHandler(authService, commentService, postService, templateService)

	// UPDATED: Post reaction handler now handles both post and comment reactions
	postReactionHandler := handlers.NewPostReactionHandler(authService, postReactionService, commentReactionService)

	// Basic routes
	mux.HandleFunc("/", homeHandler.ServeHome)
	mux.HandleFunc("/register", authHandler.ServeRegister)
	mux.HandleFunc("/login", authHandler.ServeLogin)
	mux.HandleFunc("/logout", authHandler.ServeLogout)

	// OAuth routes
	mux.HandleFunc("/auth/google/login", oauthHandler.GoogleLoginHandler)
	mux.HandleFunc("/auth/google/callback", oauthHandler.GoogleCallbackHandler)
	mux.HandleFunc("/api/auth/oauth/status", oauthHandler.OAuthStatusHandler)

	// Category routes
	mux.HandleFunc("/category/{id}", categoryHandler.ServeCategoryPosts)

	// Post routes
	mux.HandleFunc("/post/{id}", postHandler.ServePostView)
	mux.HandleFunc("/create-post", createPostHandler.ServeCreatePost)
	mux.HandleFunc("/edit-post/{id}", editPostHandler.ServeEditPost)
	mux.HandleFunc("/delete-post/{id}", deletePostHandler.ServeDeletePost)

	// Profile routes
	mux.HandleFunc("/profile", profileHandler.ServeProfile)
	mux.HandleFunc("/profile/my-posts", profileHandler.ServeUserPosts)
	mux.HandleFunc("/profile/liked-posts", profileHandler.ServeUserLikedPosts)
	mux.HandleFunc("/profile/commented-posts", profileHandler.ServeUserCommentedPosts)

	// Comment routes (form handlers)
	mux.HandleFunc("/api/comments/create/{post_id}", commentHandler.ServeCreateComment)
	mux.HandleFunc("/api/comments/edit/{comment_id}", commentHandler.ServeEditComment)
	mux.HandleFunc("/api/comments/delete/{comment_id}", commentHandler.ServeDeleteComment)
	mux.HandleFunc("/edit-comment/{id}", commentHandler.ServeEditCommentForm)
	mux.HandleFunc("/edit-comment/{id}/submit", commentHandler.ServeEditCommentSubmit)
	// Reaction routes (both post and comment reactions handled by same handler)
	mux.HandleFunc("/reactions/posts/toggle", postReactionHandler.ServeTogglePostReaction)
	mux.HandleFunc("/reactions/comments/toggle", postReactionHandler.ServeToggleCommentReaction) // NEW: Comment reactions

	return mux
}
