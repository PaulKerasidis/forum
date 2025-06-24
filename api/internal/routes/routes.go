package routes

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/PaulKerasidis/forum/config"
	"github.com/PaulKerasidis/forum/internal/handlers"
	"github.com/PaulKerasidis/forum/internal/middleware"
	"github.com/PaulKerasidis/forum/internal/repository"
)

func SetupRoutes(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	UserRepo := repository.NewUserRepository(db)
	SessionRepo := repository.NewSessionRepository(db)
	PostRepo := repository.NewPostsRepository(db)
	CategoryRepo := repository.NewCategoryRepository(db)
	CommentRepo := repository.NewCommentRepository(db)

	// NEW: Create separate reaction repositories
	PostReactionRepo := repository.NewPostReactionRepository(db)
	CommentReactionRepo := repository.NewCommentReactionRepository(db)

	AuthMiddleware := middleware.NewMiddleware(UserRepo, SessionRepo)

	// NEW: Initialize RateLimiter with config values
	RateLimiter := middleware.NewRateLimiter(
		time.Duration(config.Config.RateLimitWindow)*time.Minute,
		config.Config.RateLimitRequests,
	)

	// ===== AUTH ROUTES  =====
	mux.Handle("/api/auth/register", http.HandlerFunc(handlers.RegisterHandler(UserRepo)))
	mux.Handle("/api/auth/login", http.HandlerFunc(handlers.LoginHandler(UserRepo, SessionRepo)))
	mux.Handle("/api/auth/logout", AuthMiddleware.RequireAuth(handlers.LogoutHandler(UserRepo, SessionRepo)))
	mux.Handle("/api/auth/me", AuthMiddleware.RequireAuth(handlers.GetCurrentUser()))

	// ===== OAUTH ROUTES =====
	mux.Handle("/api/auth/google/login", http.HandlerFunc(handlers.GoogleLoginHandler()))
	mux.Handle("/api/auth/google/callback", http.HandlerFunc(handlers.GoogleCallbackHandler(UserRepo, SessionRepo)))
	mux.Handle("/api/auth/oauth/status", http.HandlerFunc(handlers.GoogleLoginStatusHandler()))

	// Add OAuth routes without /api prefix for Google callback
	mux.HandleFunc("/auth/google/login", handlers.GoogleLoginHandler())
	mux.HandleFunc("/auth/google/callback", handlers.GoogleCallbackHandler(UserRepo, SessionRepo))

	// Test route to verify routing works
	mux.HandleFunc("/auth/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OAuth routing is working!"))
	})
	// ===== USER PROFILE ROUTES =====
	mux.Handle("/api/users/profile/{id}", AuthMiddleware.RequireAuth(handlers.GetUserProfileHandler(UserRepo)))
	mux.Handle("/api/users/posts/{id}", AuthMiddleware.RequireAuth(handlers.GetUserPostsProfileHandler(PostRepo)))
	mux.Handle("/api/users/liked-posts/{id}", AuthMiddleware.RequireAuth(handlers.GetUserLikedPostsProfileHandler(PostRepo)))
	mux.Handle("/api/users/commented-posts/{id}", AuthMiddleware.RequireAuth(handlers.GetUserCommentedPostsProfileHandler(PostRepo)))

	// ===== POST ROUTES  =====
	// Public GET routes
	mux.Handle("/api/posts", http.HandlerFunc(handlers.GetAllPostsHandler(PostRepo)))
	mux.Handle("/api/posts/view/{id}", http.HandlerFunc(handlers.GetSinglePostHandler(PostRepo)))
	mux.Handle("/api/posts/by-category/{id}", http.HandlerFunc(handlers.GetPostsByCategoryHandler(PostRepo)))

	// Protected POST routes (create only)
	mux.Handle("/api/posts/create", AuthMiddleware.RequireAuth(handlers.CreatePostHandler(PostRepo, CategoryRepo)))

	// Protected PUT/DELETE routes (clear naming)
	mux.Handle("/api/posts/edit/{id}", AuthMiddleware.RequireAuth(handlers.UpdatePostHandler(PostRepo, CategoryRepo)))
	mux.Handle("/api/posts/remove/{id}", AuthMiddleware.RequireAuth(handlers.DeletePostHandler(PostRepo, CategoryRepo)))

	// ===== CATEGORY ROUTES =====
	mux.Handle("/api/categories", http.HandlerFunc(handlers.GetAllCategoriesHandler(CategoryRepo, PostRepo)))

	// ===== COMMENT ROUTES =====
	// Public GET routes
	mux.Handle("/api/comments/for-post/{id}", http.HandlerFunc(handlers.GetCommentsByPostIDHandler(CommentRepo)))

	// Protected routes
	mux.Handle("/api/comments/create-on-post/{id}", AuthMiddleware.RequireAuth(handlers.CreateCommentHandler(CommentRepo)))
	mux.Handle("/api/comments/edit/{id}", AuthMiddleware.RequireAuth(handlers.UpdateCommentHandler(CommentRepo)))
	mux.Handle("/api/comments/remove/{id}", AuthMiddleware.RequireAuth(handlers.DeleteCommentHandler(CommentRepo)))
	mux.Handle("/api/comments/view/{id}", http.HandlerFunc(handlers.GetSingleCommentHandler(CommentRepo)))

	// ===== REACTION ROUTES - UPDATED =====
	// Post reactions
	mux.Handle("/api/reactions/posts/toggle", AuthMiddleware.RequireAuth(handlers.TogglePostReactionHandler(PostReactionRepo)))

	// Comment reactions
	mux.Handle("/api/reactions/comments/toggle", AuthMiddleware.RequireAuth(handlers.ToggleCommentReactionHandler(CommentReactionRepo)))

	//...
	// Apply middleware
	handler := RateLimiter.Limit(mux)
	handler = middleware.SecurityHeaders(handler)
	handler = middleware.CORS(handler)
	return AuthMiddleware.Authenticate(handler)
}
