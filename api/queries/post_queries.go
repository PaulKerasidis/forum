package queries

const (
	// Base SELECT fields for posts -
	BaseSelectFields = `p.post_id,
		p.user_id,
		u.username,
		p.content,
		p.created_at,
		p.updated_at,
		COALESCE(like_counts.count, 0) as like_count,
		COALESCE(dislike_counts.count, 0) as dislike_count,
		COALESCE(comment_counts.count, 0) as comment_count,
		GROUP_CONCAT(DISTINCT c.category_id || ':' || c.category_name) as categories,
		ur.reaction_type as user_reaction`

	// Base JOINs for posts -
	BaseJoins = `JOIN users u ON p.user_id = u.user_id
		LEFT JOIN post_categories pc ON p.post_id = pc.post_id
		LEFT JOIN categories c ON pc.category_id = c.category_id`

	// Category filtering JOIN -
	CategoryFilterJoin = `JOIN users u ON p.user_id = u.user_id
		JOIN post_categories pc ON p.post_id = pc.post_id
		LEFT JOIN categories c ON pc.category_id = c.category_id`

	// Liked posts filtering JOIN - UPDATED to use post_reactions
	LikedPostsJoin = `JOIN post_reactions r ON p.post_id = r.post_id
		JOIN users u ON p.user_id = u.user_id
		LEFT JOIN post_categories pc ON p.post_id = pc.post_id
		LEFT JOIN categories c ON pc.category_id = c.category_id`

	// Commented posts filtering JOIN -
	CommentedPostsJoin = `JOIN comments com ON p.post_id = com.post_id
		JOIN users u ON p.user_id = u.user_id
		LEFT JOIN post_categories pc ON p.post_id = pc.post_id
		LEFT JOIN categories c ON pc.category_id = c.category_id`

	// Reaction count subqueries - UPDATED to use post_reactions table
	ReactionCountJoins = `LEFT JOIN (
			SELECT post_id, COUNT(*) as count 
			FROM post_reactions 
			WHERE reaction_type = 1
			GROUP BY post_id
		) like_counts ON p.post_id = like_counts.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) as count 
			FROM post_reactions 
			WHERE reaction_type = 2
			GROUP BY post_id
		) dislike_counts ON p.post_id = dislike_counts.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) as count
			FROM comments
			GROUP BY post_id
		) comment_counts ON p.post_id = comment_counts.post_id`

	// User reaction JOIN - UPDATED to use post_reactions table
	UserReactionJoin = `LEFT JOIN post_reactions ur ON p.post_id = ur.post_id AND ur.user_id = ?`

	// Common clauses -
	GroupByPost          = `GROUP BY p.post_id`
	OrderByCreated       = `ORDER BY p.created_at DESC`
	OrderByLikedDate     = `ORDER BY r.created_at DESC`
	OrderByCommentedDate = `ORDER BY MAX(com.created_at) DESC`
	LimitOffset          = `LIMIT ? OFFSET ?`

	// Base WHERE clause for dynamic filtering -
	BaseWhere = `WHERE 1=1`
)

// Static queries -
var (
	GetPostByIDQuery = `SELECT ` + BaseSelectFields + `
		FROM posts p
		` + BaseJoins + `
		` + ReactionCountJoins + `
		` + UserReactionJoin + `
		WHERE p.post_id = ?
		` + GroupByPost

	GetAllPostsQuery = `SELECT ` + BaseSelectFields + `
		FROM posts p
		` + BaseJoins + `
		` + ReactionCountJoins + `
		` + UserReactionJoin + `
		` + GroupByPost + `
		` + OrderByCreated + `
		` + LimitOffset

	GetPostsByCategoryQuery = `SELECT ` + BaseSelectFields + `
		FROM posts p
		` + CategoryFilterJoin + `
		` + ReactionCountJoins + `
		` + UserReactionJoin + `
		WHERE pc.category_id = ?
		` + GroupByPost + `
		` + OrderByCreated + `
		` + LimitOffset

	GetPostsByUserQuery = `SELECT ` + BaseSelectFields + `
		FROM posts p
		` + BaseJoins + `
		` + ReactionCountJoins + `
		` + UserReactionJoin + `
		WHERE p.user_id = ?
		` + GroupByPost + `
		` + OrderByCreated + `
		` + LimitOffset

	GetPostsLikedByUserQuery = `SELECT ` + BaseSelectFields + `
		FROM posts p
		` + LikedPostsJoin + `
		` + ReactionCountJoins + `
		` + UserReactionJoin + `
		WHERE r.user_id = ? AND r.reaction_type = 1
		` + GroupByPost + `
		` + OrderByLikedDate + `
		` + LimitOffset

	GetPostsCommentedByUserQuery = `SELECT ` + BaseSelectFields + `
		FROM posts p
		` + CommentedPostsJoin + `
		` + ReactionCountJoins + `
		` + UserReactionJoin + `
		WHERE com.user_id = ?
		` + GroupByPost + `
		` + OrderByCommentedDate + `
		` + LimitOffset
)

// NEW: Dynamic Query Builder Functions

// BuildPostsQuery creates a dynamic query with sorting and filtering options
func BuildPostsQuery(joins, whereClause, orderClause string) string {
	return `SELECT ` + BaseSelectFields + `
		FROM posts p
		` + joins + `
		` + ReactionCountJoins + `
		` + UserReactionJoin + `
		` + whereClause + `
		` + GroupByPost + `
		` + orderClause + `
		` + LimitOffset
}

// GetAllPostsWithSortQuery returns a dynamic query for all posts with custom sorting
func GetAllPostsWithSortQuery(orderClause string) string {
	return BuildPostsQuery(BaseJoins, BaseWhere, orderClause)
}

// GetPostsByCategoryWithSortQuery returns a dynamic query for category posts with custom sorting
func GetPostsByCategoryWithSortQuery(orderClause string) string {
	whereClause := `WHERE pc.category_id = ?`
	return BuildPostsQuery(CategoryFilterJoin, whereClause, orderClause)
}

// GetPostsByUserWithSortQuery returns a dynamic query for user posts with custom sorting
func GetPostsByUserWithSortQuery(orderClause string) string {
	whereClause := `WHERE p.user_id = ?`
	return BuildPostsQuery(BaseJoins, whereClause, orderClause)
}

// GetPostsLikedByUserWithSortQuery returns a dynamic query for liked posts with custom sorting
func GetPostsLikedByUserWithSortQuery(orderClause string) string {
	whereClause := `WHERE r.user_id = ? AND r.reaction_type = 1`
	return BuildPostsQuery(LikedPostsJoin, whereClause, orderClause)
}

// GetPostsCommentedByUserWithSortQuery returns a dynamic query for commented posts with custom sorting
func GetPostsCommentedByUserWithSortQuery(orderClause string) string {
	whereClause := `WHERE com.user_id = ?`
	return BuildPostsQuery(CommentedPostsJoin, whereClause, orderClause)
}
