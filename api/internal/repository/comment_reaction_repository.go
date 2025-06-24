package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/PaulKerasidis/forum/internal/models"
	"github.com/PaulKerasidis/forum/internal/utils"
)

type CommentReactionRepository struct {
	db *sql.DB
}

// NewCommentReactionRepository creates a new CommentReactionRepository
func NewCommentReactionRepository(db *sql.DB) *CommentReactionRepository {
	return &CommentReactionRepository{db: db}
}

// ToggleCommentReaction handles like/dislike toggle logic for comments
func (crr *CommentReactionRepository) ToggleCommentReaction(userID, commentID string, reactionType int) (*models.ReactionResult, error) {
	return utils.ExecuteInTransactionWithResult(crr.db, func(tx *sql.Tx) (*models.ReactionResult, error) {
		// Validate that the comment exists
		if err := crr.validateCommentExists(tx, commentID); err != nil {
			return nil, err
		}

		// Check if a reaction already exists
		var existingType sql.NullInt32
		err := tx.QueryRow("SELECT reaction_type FROM comment_reactions WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&existingType)

		var action string

		if err == sql.ErrNoRows {
			// No existing reaction - CREATE new reaction
			action, err = crr.createCommentReaction(tx, userID, commentID, reactionType)
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			// Database error
			return nil, err
		} else {
			// Reaction exists - determine what to do
			existingReactionType := int(existingType.Int32)

			if existingReactionType == reactionType {
				// Same reaction - REMOVE it (toggle off)
				action, err = crr.removeCommentReaction(tx, userID, commentID, reactionType)
				if err != nil {
					return nil, err
				}
			} else {
				// Different reaction - CHANGE it
				action, err = crr.changeCommentReaction(tx, userID, commentID, existingReactionType, reactionType)
				if err != nil {
					return nil, err
				}
			}
		}

		// Return the result with action and message
		return models.NewReactionResult(action), nil
	})
}

// Helper method to create a new comment reaction
func (crr *CommentReactionRepository) createCommentReaction(tx *sql.Tx, userID, commentID string, reactionType int) (string, error) {
	createdAt := time.Now()

	_, err := tx.Exec(
		"INSERT INTO comment_reactions (user_id, comment_id, reaction_type, created_at) VALUES (?, ?, ?, ?)",
		userID, commentID, reactionType, createdAt,
	)
	if err != nil {
		return "", err
	}

	// Return appropriate action
	if reactionType == models.ReactionTypeLike {
		return models.ActionCommentLikeCreated, nil
	}
	return models.ActionCommentDislikeCreated, nil
}

// Helper method to remove an existing comment reaction
func (crr *CommentReactionRepository) removeCommentReaction(tx *sql.Tx, userID, commentID string, reactionType int) (string, error) {
	_, err := tx.Exec("DELETE FROM comment_reactions WHERE user_id = ? AND comment_id = ?", userID, commentID)
	if err != nil {
		return "", err
	}

	// Return appropriate action
	if reactionType == models.ReactionTypeLike {
		return models.ActionCommentLikeRemoved, nil
	}
	return models.ActionCommentDislikeRemoved, nil
}

// Helper method to change an existing comment reaction
func (crr *CommentReactionRepository) changeCommentReaction(tx *sql.Tx, userID, commentID string, existingType, newType int) (string, error) {
	_, err := tx.Exec("UPDATE comment_reactions SET reaction_type = ? WHERE user_id = ? AND comment_id = ?", newType, userID, commentID)
	if err != nil {
		return "", err
	}

	// Return appropriate action
	if existingType == models.ReactionTypeLike && newType == models.ReactionTypeDislike {
		return models.ActionCommentLikeToDislike, nil
	}
	return models.ActionCommentDislikeToLike, nil
}

// Helper method to validate that a comment exists
func (crr *CommentReactionRepository) validateCommentExists(tx *sql.Tx, commentID string) error {
	var exists int
	err := tx.QueryRow("SELECT COUNT(*) FROM comments WHERE comment_id = ?", commentID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return errors.New("comment not found")
	}
	return nil
}
