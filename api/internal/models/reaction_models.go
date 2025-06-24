package models

// Reaction type constants
const (
	ReactionTypeLike    = 1
	ReactionTypeDislike = 2
)

// Reaction action constants - ALL reactions
const (
	// Post reactions
	ActionPostLikeCreated    = "post_like_created"
	ActionPostDislikeCreated = "post_dislike_created"
	ActionPostLikeRemoved    = "post_like_removed"
	ActionPostDislikeRemoved = "post_dislike_removed"
	ActionPostLikeToDislike  = "post_like_to_dislike"
	ActionPostDislikeToLike  = "post_dislike_to_like"

	// Comment reactions
	ActionCommentLikeCreated    = "comment_like_created"
	ActionCommentDislikeCreated = "comment_dislike_created"
	ActionCommentLikeRemoved    = "comment_like_removed"
	ActionCommentDislikeRemoved = "comment_dislike_removed"
	ActionCommentLikeToDislike  = "comment_like_to_dislike"
	ActionCommentDislikeToLike  = "comment_dislike_to_like"
)

// ReactionResult represents the outcome of a reaction toggle operation
type ReactionResult struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

// PostReaction - Database model for post_reactions table
type PostReaction struct {
	UserID       string `json:"user_id"`
	PostID       string `json:"post_id"`
	ReactionType int    `json:"reaction_type"`
	CreatedAt    string `json:"created_at"`
}

// CommentReaction - Database model for comment_reactions table
type CommentReaction struct {
	UserID       string `json:"user_id"`
	CommentID    string `json:"comment_id"`
	ReactionType int    `json:"reaction_type"`
	CreatedAt    string `json:"created_at"`
}

// IsValidReactionType checks if the reaction type is valid
func IsValidReactionType(reactionType int) bool {
	return reactionType == ReactionTypeLike || reactionType == ReactionTypeDislike
}

// Helper function to create ReactionResult with appropriate message
func NewReactionResult(action string) *ReactionResult {
	messages := map[string]string{
		// Post reactions
		ActionPostLikeCreated:    "You liked this post",
		ActionPostDislikeCreated: "You disliked this post",
		ActionPostLikeRemoved:    "You removed your like from this post",
		ActionPostDislikeRemoved: "You removed your dislike from this post",
		ActionPostLikeToDislike:  "You changed your like to dislike on this post",
		ActionPostDislikeToLike:  "You changed your dislike to like on this post",

		// Comment reactions
		ActionCommentLikeCreated:    "You liked this comment",
		ActionCommentDislikeCreated: "You disliked this comment",
		ActionCommentLikeRemoved:    "You removed your like from this comment",
		ActionCommentDislikeRemoved: "You removed your dislike from this comment",
		ActionCommentLikeToDislike:  "You changed your like to dislike on this comment",
		ActionCommentDislikeToLike:  "You changed your dislike to like on this comment",
	}

	return &ReactionResult{
		Action:  action,
		Message: messages[action],
	}
}
