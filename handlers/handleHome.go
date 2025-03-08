package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

// Forum data structures
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Post struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
	CategoryID int    `json:"category_id"`
	Category   string `json:"category"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at,omitempty"`
	Likes      int    `json:"likes"`
	Dislikes   int    `json:"dislikes"`
}

type HomePageData struct {
	Categories []Category `json:"categories"`
	Posts      []Post     `json:"posts"`
}

// HomeHandler handles requests to the home page ("/")
func HomeHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Check if the requested URL path is "/" (the home page)
	if r.URL.Path != "/" {
		// If not, return a 404 Not Found response with JSON error
		sendJSONError(w, http.StatusNotFound, "Page Not Found")
		return
	}

	// Prepare data for the response
	data, err := getHomePageData(db)
	if err != nil {
		log.Printf("Error fetching home page data: %v", err)
		sendJSONError(w, http.StatusInternalServerError, "Failed to load content")
		return
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Encode and send the JSON response
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("JSON encoding error: %v", err)
		sendJSONError(w, http.StatusInternalServerError, "Could not encode response")
		return
	}
}

// getHomePageData fetches all categories and posts with their like/dislike counts
func getHomePageData(db *sql.DB) (HomePageData, error) {
	var data HomePageData

	// 1. Fetch all categories
	categories, err := fetchAllCategories(db)
	if err != nil {
		return data, err
	}
	data.Categories = categories

	// 2. Fetch all posts with their metadata
	posts, err := fetchAllPosts(db)
	if err != nil {
		return data, err
	}
	data.Posts = posts

	return data, nil
}

// fetchAllCategories gets all categories from the database
func fetchAllCategories(db *sql.DB) ([]Category, error) {
	query := `SELECT category_id, name FROM categories ORDER BY name`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// fetchAllPosts gets all posts with user info, category name, and reaction counts
func fetchAllPosts(db *sql.DB) ([]Post, error) {
	query := `
		SELECT 
			p.post_id, 
			p.user_id, 
			u.username, 
			p.category_id, 
			c.name AS category_name, 
			p.content, 
			p.created_at, 
			p.updated_at,
			(SELECT COUNT(*) FROM reactions WHERE post_id = p.post_id AND reaction_type = 1) AS likes,
			(SELECT COUNT(*) FROM reactions WHERE post_id = p.post_id AND reaction_type = 2) AS dislikes
		FROM posts p
		JOIN user u ON p.user_id = u.user_id
		JOIN categories c ON p.category_id = c.category_id
		ORDER BY p.created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var updatedAt sql.NullString

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Username,
			&post.CategoryID,
			&post.Category,
			&post.Content,
			&post.CreatedAt,
			&updatedAt,
			&post.Likes,
			&post.Dislikes,
		)
		if err != nil {
			return nil, err
		}

		if updatedAt.Valid {
			post.UpdatedAt = updatedAt.String
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
