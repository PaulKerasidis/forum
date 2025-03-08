package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// CategoryFilterResponse represents the JSON response for category filtering
type CategoryFilterResponse struct {
	CategoryID   int    `json:"categoryId"`
	CategoryName string `json:"categoryName"`
	Posts        []Post `json:"posts"`
}

// CategoryFilterHandler handles requests to filter posts by category ID
func CategoryFilterHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Extract category ID from URL path
	// The path will be like /categories/1
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		sendJSONError(w, http.StatusBadRequest, "Invalid URL format")
		return
	}

	// Get the category ID from the path
	categoryIDStr := pathParts[2]
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		log.Printf("Invalid category ID format: %v", err)
		sendJSONError(w, http.StatusBadRequest, "Invalid category ID format")
		return
	}

	// Get category name
	var categoryName string
	err = db.QueryRow("SELECT name FROM categories WHERE category_id = ?", categoryID).Scan(&categoryName)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJSONError(w, http.StatusNotFound, "Category not found")
		} else {
			log.Printf("Database error fetching category: %v", err)
			sendJSONError(w, http.StatusInternalServerError, "Database error")
		}
		return
	}

	// Fetch posts for this category
	posts, err := fetchPostsByCategoryID(db, categoryID)
	if err != nil {
		log.Printf("Error fetching posts for category %d: %v", categoryID, err)
		sendJSONError(w, http.StatusInternalServerError, "Error fetching posts")
		return
	}

	// Prepare the response
	response := CategoryFilterResponse{
		CategoryID:   categoryID,
		CategoryName: categoryName,
		Posts:        posts,
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		sendJSONError(w, http.StatusInternalServerError, "Error generating response")
		return
	}

	log.Printf("Successfully returned %d posts for category %d (%s)", len(posts), categoryID, categoryName)
}

// fetchPostsByCategoryID retrieves all posts for a specific category ID
func fetchPostsByCategoryID(db *sql.DB, categoryID int) ([]Post, error) {
	query := `
        SELECT 
            p.post_id, 
            p.user_id, 
            u.username, 
            p.category_id, 
            c.name, 
            p.content, 
            p.created_at, 
            p.updated_at,
            (SELECT COUNT(*) FROM reactions WHERE post_id = p.post_id AND reaction_type = 1) AS likes,
            (SELECT COUNT(*) FROM reactions WHERE post_id = p.post_id AND reaction_type = 2) AS dislikes
        FROM 
            posts p
        JOIN 
            user u ON p.user_id = u.user_id
        JOIN 
            categories c ON p.category_id = c.category_id
        WHERE 
            p.category_id = ?
        ORDER BY 
            p.created_at DESC
    `

	rows, err := db.Query(query, categoryID)
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

	// It's perfectly fine to return an empty slice if no posts are found
	if posts == nil {
		posts = []Post{}
	}

	return posts, nil
}
