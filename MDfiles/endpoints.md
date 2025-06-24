POST /api/auth/register      # Register new user
POST /api/auth/login         # Login user (creates session)
POST /api/auth/logout        # Logout user (requires auth)
POST /api/auth/me            # Get current user info (requires auth)

#  Post crud operations
GET  /api/posts                    # Get all posts with pagination
GET  /api/posts/{id}              # Get single post (no comments)
POST /api/posts                   # Create new post (requires auth)
PUT  /api/posts/{id}              # Update post (requires auth + ownership)
DELETE /api/posts/{id}            # Delete post (requires auth + ownership)

# Post filtering
GET /api/posts/category/{id}      # Get posts by category
GET /api/posts/user/{id}          # Get posts by user
GET /api/posts/liked              # Get current user's liked posts (requires auth)

# Comments Endpoints
GET  /api/posts/{id}/comments     # Get comments for a post (paginated)
POST /api/posts/{id}/comments     # Create comment on post (requires auth)
PUT  /api/comments/{id}           # Update comment (requires auth + ownership)
DELETE /api/comments/{id}         # Delete comment (requires auth + ownership)

# Reactions Endpoints
POST /api/reactions/posts/toggle    # Toggle like/dislike on post (requires auth)
POST /api/reactions/comments/toggle # Toggle like/dislike on comment (requires auth)

# Categories Endpoints
GET /api/categories               # Get all categories with post counts

# Profile Endpoints
GET /api/users/{id}               # Get user profile
GET /api/users/{id}/posts         # Get user's posts (paginated)
GET /api/users/{id}/comments      # Get user's comments (paginated)