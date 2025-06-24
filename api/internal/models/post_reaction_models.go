package models

import "errors"

// PostReactionRequest - Request payload for post reactions
type PostReactionRequest struct {
	PostID       string `json:"post_id" binding:"required"`
	ReactionType int    `json:"reaction_type" binding:"required,min=1,max=2"`
}

// Validate validates the post reaction request
func (req *PostReactionRequest) Validate() error {
	if req.PostID == "" {
		return errors.New("post_id is required and cannot be empty")
	}
	if !IsValidReactionType(req.ReactionType) {
		return errors.New("invalid reaction type. Use 1 for like, 2 for dislike")
	}
	return nil
}
