package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/PaulKerasidis/forum/internal/models"
	"github.com/PaulKerasidis/forum/internal/utils"
	"github.com/PaulKerasidis/forum/queries"
)

type PostsRepository struct {
	db *sql.DB
}

// NewPostsRepository creates a new PostsRepository
func NewPostsRepository(db *sql.DB) *PostsRepository {
	return &PostsRepository{db: db}
}

// CRUD methods
func (pr *PostsRepository) CreatePost(userID string, content string, categoryIDs []string) (*models.CreatePostResponse, error) {
	return utils.ExecuteInTransactionWithResult(pr.db, func(tx *sql.Tx) (*models.CreatePostResponse, error) {
		// Generate UUID for the post
		postID := utils.GenerateUUIDToken()
		createdAt := time.Now()

		// Insert post
		_, err := tx.Exec(
			"INSERT INTO posts (post_id, user_id, content, created_at) VALUES (?, ?, ?, ?)",
			postID, userID, content, createdAt,
		)
		if err != nil {
			return nil, err
		}

		// Insert post-category associations
		for _, categoryID := range categoryIDs {
			_, err := tx.Exec(
				"INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)",
				postID, categoryID,
			)
			if err != nil {
				return nil, err
			}
		}

		// Return lightweight response - just ID and timestamp
		return &models.CreatePostResponse{
			PostID:    postID,
			CreatedAt: createdAt,
		}, nil
	})
}

func (pr *PostsRepository) UpdatePost(postID, userID, content string, categoryIDs []string) error {
	return utils.ExecuteInTransaction(pr.db, func(tx *sql.Tx) error {
		// Check if user owns the post
		var ownerID string
		err := tx.QueryRow("SELECT user_id FROM posts WHERE post_id = ?", postID).Scan(&ownerID)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.New("post not found")
			}
			return err
		}

		if ownerID != userID {
			return errors.New("unauthorized: you can only update your own posts")
		}

		now := time.Now()

		// 1. Update post content AND set updated_at
		_, err = tx.Exec("UPDATE posts SET content = ?, updated_at = ? WHERE post_id = ?", content, now, postID)
		if err != nil {
			return err
		}

		// 2. UPDATE CATEGORIES - Delete existing categories first
		_, err = tx.Exec("DELETE FROM post_categories WHERE post_id = ?", postID)
		if err != nil {
			return err
		}

		// 3. INSERT NEW CATEGORIES
		for _, categoryID := range categoryIDs {
			_, err := tx.Exec(
				"INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)",
				postID, categoryID,
			)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (pr *PostsRepository) DeletePost(postID, userID string) error {
	return utils.ExecuteInTransaction(pr.db, func(tx *sql.Tx) error {
		var ownerID string
		err := tx.QueryRow("SELECT user_id FROM posts WHERE post_id = ?", postID).Scan(&ownerID)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.New("post not found")
			}
			return err
		}

		if ownerID != userID {
			return errors.New("unauthorized: you can only delete your own posts")
		}

		// Delete the post (CASCADE will handle related records)
		_, err = tx.Exec("DELETE FROM posts WHERE post_id = ?", postID)
		if err != nil {
			return err
		}

		return nil
	})
}

// GET METHODS - Using Queries Package

func (pr *PostsRepository) GetPostByID(postID string, userID *string) (*models.Post, error) {
	// Prepare arguments
	var userIDArg interface{}
	if userID != nil {
		userIDArg = *userID
	} else {
		userIDArg = "" // Won't match any user_id
	}

	row := pr.db.QueryRow(queries.GetPostByIDQuery, userIDArg, postID)
	post, err := pr.scanAndParsePost(row, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("post not found")
		}
		return nil, err
	}

	return post, nil
}

// GetAllPosts retrieves all posts (sorted by newest first by default, or custom sorting)
func (pr *PostsRepository) GetAllPosts(limit, offset int, userID *string, options utils.SortOptions) ([]*models.Post, error) {
	// Prepare arguments
	var userIDArg interface{}
	if userID != nil {
		userIDArg = *userID
	} else {
		userIDArg = "" // Won't match any user_id
	}

	// Build dynamic query with sort options
	orderClause := utils.BuildOrderClause(options.SortBy, utils.ContentTypePosts)
	query := queries.GetAllPostsWithSortQuery(orderClause)

	rows, err := pr.db.Query(query, userIDArg, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post, err := pr.scanAndParsePost(rows, userID)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// GetPostsByCategory retrieves posts by category (sorted by newest first by default, or custom sorting)
func (pr *PostsRepository) GetPostsByCategory(categoryID string, limit, offset int, userID *string, options utils.SortOptions) ([]*models.Post, error) {
	// Prepare arguments
	var userIDArg interface{}
	if userID != nil {
		userIDArg = *userID
	} else {
		userIDArg = "" // Won't match any user_id
	}

	// Build dynamic query with sort options
	orderClause := utils.BuildOrderClause(options.SortBy, utils.ContentTypePosts)
	query := queries.GetPostsByCategoryWithSortQuery(orderClause)

	rows, err := pr.db.Query(query, userIDArg, categoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post, err := pr.scanAndParsePost(rows, userID)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// ...
// Profiling methods for user-specific posts
// ...
// GetPostsByUser retrieves posts by user (sorted by newest first by default, or custom sorting)
func (pr *PostsRepository) GetPostsByUser(targetUserID string, limit, offset int, userID *string, options utils.SortOptions) ([]*models.Post, error) {
	// Prepare arguments
	var userIDArg interface{}
	if userID != nil {
		userIDArg = *userID
	} else {
		userIDArg = "" // Won't match any user_id
	}

	// Build dynamic query with sort options
	orderClause := utils.BuildOrderClause(options.SortBy, utils.ContentTypePosts)
	query := queries.GetPostsByUserWithSortQuery(orderClause)

	rows, err := pr.db.Query(query, userIDArg, targetUserID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post, err := pr.scanAndParsePost(rows, userID)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// GetPostsLikedByUser retrieves posts liked by user (sorted by newest first by default, or custom sorting)
func (pr *PostsRepository) GetPostsLikedByUser(targetUserID string, limit, offset int, userID *string, options utils.SortOptions) ([]*models.Post, error) {
	// Prepare arguments
	var userIDArg interface{}
	if userID != nil {
		userIDArg = *userID
	} else {
		userIDArg = "" // Won't match any user_id
	}

	// Build dynamic query with sort options
	orderClause := utils.BuildOrderClause(options.SortBy, utils.ContentTypePosts)
	query := queries.GetPostsLikedByUserWithSortQuery(orderClause)

	rows, err := pr.db.Query(query, userIDArg, targetUserID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post, err := pr.scanAndParsePost(rows, userID)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// GetPostsCommentedByUser retrieves posts commented by user (sorted by newest first by default, or custom sorting)
func (pr *PostsRepository) GetPostsCommentedByUser(targetUserID string, limit, offset int, userID *string, options utils.SortOptions) ([]*models.Post, error) {
	// Prepare arguments
	var userIDArg interface{}
	if userID != nil {
		userIDArg = *userID
	} else {
		userIDArg = "" // Won't match any user_id
	}

	// Build dynamic query with sort options
	orderClause := utils.BuildOrderClause(options.SortBy, utils.ContentTypePosts)
	query := queries.GetPostsCommentedByUserWithSortQuery(orderClause)

	rows, err := pr.db.Query(query, userIDArg, targetUserID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post, err := pr.scanAndParsePost(rows, userID)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// HELPER METHODS FOR POST COUNT
// get all the post count
func (pr *PostsRepository) GetCountTotalPosts() (int, error) {
	var count int
	err := pr.db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&count)
	return count, err
}

// get all the post count by category
func (pr *PostsRepository) GetCountPostByCategory(categoryID string) (int, error) {
	var count int
	err := pr.db.QueryRow("SELECT COUNT(*) FROM post_categories WHERE category_id = ?", categoryID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil

}

// Get post count created by the user
func (pr *PostsRepository) GetCountPostByUser(userID string) (int, error) {
	var count int
	err := pr.db.QueryRow("SELECT COUNT(*) FROM posts WHERE user_id = ?", userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil

}

// GetCommentedPostCountByUser returns the total number of posts a user has commented on
func (pr *PostsRepository) GetCountCommentedPostByUser(userID string) (int, error) {
	var count int
	err := pr.db.QueryRow(`
		SELECT COUNT(DISTINCT p.post_id) 
		FROM posts p
		JOIN comments c ON p.post_id = c.post_id
		WHERE c.user_id = ?
	`, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetLikedPostCountByUser returns the total number of posts liked by a specific user
func (pr *PostsRepository) GetCountLikedPostByUser(userID string) (int, error) {
	var count int
	// FIXED: use post_reactions table instead of non-existent reactions table
	err := pr.db.QueryRow(`
		SELECT COUNT(*) 
		FROM post_reactions 
		WHERE user_id = ? AND reaction_type = 1
	`, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

//...
// Helper method to parse categories , handle updated_at and user reaction
//...

// Helper method to scan and parse a single post row
func (pr *PostsRepository) scanAndParsePost(rows interface{}, userID *string) (*models.Post, error) {
	var post models.Post
	var categoriesStr sql.NullString
	var userReaction sql.NullInt64
	var updatedAt sql.NullTime

	var err error

	// Handle both *sql.Row and *sql.Rows
	switch v := rows.(type) {
	case *sql.Row:
		err = v.Scan(
			&post.ID,
			&post.UserID,
			&post.Username,
			&post.Content,
			&post.CreatedAt,
			&updatedAt,
			&post.LikeCount,
			&post.DislikeCount,
			&post.CommentCount,
			&categoriesStr,
			&userReaction,
		)
	case *sql.Rows:
		err = v.Scan(
			&post.ID,
			&post.UserID,
			&post.Username,
			&post.Content,
			&post.CreatedAt,
			&updatedAt,
			&post.LikeCount,
			&post.DislikeCount,
			&post.CommentCount,
			&categoriesStr,
			&userReaction,
		)
	default:
		return nil, errors.New("invalid row type")
	}

	if err != nil {
		return nil, err
	}

	// Parse categories directly
	if categoriesStr.Valid && categoriesStr.String != "" {
		categoryPairs := strings.Split(categoriesStr.String, ",")
		for _, pair := range categoryPairs {
			parts := strings.Split(strings.TrimSpace(pair), ":")
			if len(parts) == 2 {
				post.Categories = append(post.Categories, models.PostCategory{
					ID:   parts[0],
					Name: parts[1],
				})
			}
		}
	}

	// Handle UpdatedAt directly
	if updatedAt.Valid {
		post.UpdatedAt = &updatedAt.Time
	} else {
		post.UpdatedAt = nil
	}

	// Handle UserReaction - NO CHANGE NEEDED
	if userReaction.Valid {
		reactionType := int(userReaction.Int64)
		post.UserReaction = &reactionType
	} else {
		post.UserReaction = nil
	}

	// Handle IsOwner - NO CHANGE NEEDED
	post.IsOwner = (userID != nil && post.UserID == *userID)

	return &post, nil
}
