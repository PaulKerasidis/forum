package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/PaulKerasidis/forum/internal/models"
	"github.com/PaulKerasidis/forum/internal/utils"
)

type PostReactionRepository struct {
	db *sql.DB
}

// NewPostReactionRepository creates a new PostReactionRepository
func NewPostReactionRepository(db *sql.DB) *PostReactionRepository {
	return &PostReactionRepository{db: db}
}

// TogglePostReaction handles like/dislike toggle logic for posts
func (prr *PostReactionRepository) TogglePostReaction(userID, postID string, reactionType int) (*models.ReactionResult, error) {
	return utils.ExecuteInTransactionWithResult(prr.db, func(tx *sql.Tx) (*models.ReactionResult, error) {
		// Validate that the post exists
		if err := prr.validatePostExists(tx, postID); err != nil {
			return nil, err
		}

		// Check if a reaction already exists
		var existingType sql.NullInt32
		err := tx.QueryRow("SELECT reaction_type FROM post_reactions WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingType)

		var action string

		if err == sql.ErrNoRows {
			// No existing reaction - CREATE new reaction
			action, err = prr.createPostReaction(tx, userID, postID, reactionType)
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
				action, err = prr.removePostReaction(tx, userID, postID, reactionType)
				if err != nil {
					return nil, err
				}
			} else {
				// Different reaction - CHANGE it
				action, err = prr.changePostReaction(tx, userID, postID, existingReactionType, reactionType)
				if err != nil {
					return nil, err
				}
			}
		}

		// Return the result with action and message
		return models.NewReactionResult(action), nil
	})
}

// Helper method to create a new post reaction
func (prr *PostReactionRepository) createPostReaction(tx *sql.Tx, userID, postID string, reactionType int) (string, error) {
	createdAt := time.Now()

	_, err := tx.Exec(
		"INSERT INTO post_reactions (user_id, post_id, reaction_type, created_at) VALUES (?, ?, ?, ?)",
		userID, postID, reactionType, createdAt,
	)
	if err != nil {
		return "", err
	}

	// Return appropriate action
	if reactionType == models.ReactionTypeLike {
		return models.ActionPostLikeCreated, nil
	}
	return models.ActionPostDislikeCreated, nil
}

// Helper method to remove an existing post reaction
func (prr *PostReactionRepository) removePostReaction(tx *sql.Tx, userID, postID string, reactionType int) (string, error) {
	_, err := tx.Exec("DELETE FROM post_reactions WHERE user_id = ? AND post_id = ?", userID, postID)
	if err != nil {
		return "", err
	}

	// Return appropriate action
	if reactionType == models.ReactionTypeLike {
		return models.ActionPostLikeRemoved, nil
	}
	return models.ActionPostDislikeRemoved, nil
}

// Helper method to change an existing post reaction
func (prr *PostReactionRepository) changePostReaction(tx *sql.Tx, userID, postID string, existingType, newType int) (string, error) {
	_, err := tx.Exec("UPDATE post_reactions SET reaction_type = ? WHERE user_id = ? AND post_id = ?", newType, userID, postID)
	if err != nil {
		return "", err
	}

	// Return appropriate action
	if existingType == models.ReactionTypeLike && newType == models.ReactionTypeDislike {
		return models.ActionPostLikeToDislike, nil
	}
	return models.ActionPostDislikeToLike, nil
}

// Helper method to validate that a post exists
func (prr *PostReactionRepository) validatePostExists(tx *sql.Tx, postID string) error {
	var exists int
	err := tx.QueryRow("SELECT COUNT(*) FROM posts WHERE post_id = ?", postID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return errors.New("post not found")
	}
	return nil
}
