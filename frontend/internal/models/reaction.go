package models

// Reaction type constants (matches backend exactly)
const (
	ReactionTypeLike    = 1
	ReactionTypeDislike = 2
)

// ReactionResult - Result of a reaction toggle operation (matches backend exactly)
type ReactionResult struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

// PostReactionRequest - Request to toggle post reaction (matches backend exactly)
type PostReactionRequest struct {
	PostID       string `json:"post_id"`
	ReactionType int    `json:"reaction_type"`
}

// CommentReactionRequest - Request to toggle comment reaction (matches backend exactly)
type CommentReactionRequest struct {
	CommentID    string `json:"comment_id"`
	ReactionType int    `json:"reaction_type"`
}
