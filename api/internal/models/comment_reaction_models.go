package models

import "errors"

// CommentReactionRequest - Request payload for comment reactions
type CommentReactionRequest struct {
	CommentID    string `json:"comment_id" binding:"required"`
	ReactionType int    `json:"reaction_type" binding:"required,min=1,max=2"`
}

// Validate validates the comment reaction request
func (req *CommentReactionRequest) Validate() error {
	if req.CommentID == "" {
		return errors.New("comment_id is required and cannot be empty")
	}
	if !IsValidReactionType(req.ReactionType) {
		return errors.New("invalid reaction type. Use 1 for like, 2 for dislike")
	}
	return nil
}
