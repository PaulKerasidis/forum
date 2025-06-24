package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/PaulKerasidis/forum/internal/models"
	"github.com/PaulKerasidis/forum/internal/utils"
)

// UserRepository handles user-related database operations
type UserRepository struct {
	DB *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (ur *UserRepository) CreateUser(reg models.UserRegistration) (*models.User, error) {
	return utils.ExecuteInTransactionWithResult(ur.DB, func(tx *sql.Tx) (*models.User, error) {
		// Check if username exists
		var usernameCount int
		err := tx.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", reg.Username).Scan(&usernameCount)
		if err != nil {
			return nil, err
		}
		if usernameCount > 0 {
			return nil, errors.New("username already taken")
		}

		// Check if email exists
		var emailCount int
		err = tx.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", reg.Email).Scan(&emailCount)
		if err != nil {
			return nil, err
		}
		if emailCount > 0 {
			return nil, errors.New("email already taken")
		}

		userID := utils.GenerateUUIDToken()
		createdAt := time.Now()

		// Hash the password
		hashedPassword, err := utils.HashPassword(reg.Password)
		if err != nil {
			return nil, err
		}

		// Insert user record
		_, err = tx.Exec(
			"INSERT INTO users (user_id, username, email, password_hash, created_at) VALUES (?, ?, ?, ?, ?)",
			userID, reg.Username, reg.Email, hashedPassword, createdAt,
		)
		if err != nil {
			return nil, err
		}

		// Return the created user
		return &models.User{
			ID:        userID,
			Username:  reg.Username,
			Email:     reg.Email,
			CreatedAt: createdAt,
		}, nil
	})
}

// GetBySessionID retrieves a user by session ID
func (ur *UserRepository) GetUserBySessionID(id string) (*models.User, error) {
	var user models.User

	err := ur.DB.QueryRow(
		"SELECT user_id, username, email, COALESCE(provider, ''), COALESCE(provider_id, ''), COALESCE(provider_email, ''), created_at FROM users WHERE user_id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Provider, &user.ProviderID, &user.ProviderEmail, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil

}

// GetAuthByUserID retrieves user authentication data by user ID
func (ur *UserRepository) GetAuthByUserID(userID string) (*models.UserPassword, error) {
	var auth models.UserPassword

	err := ur.DB.QueryRow(
		"SELECT user_id, password_hash FROM users WHERE user_id = ?",
		userID,
	).Scan(&auth.UserID, &auth.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user authentication not found")
		}
		return nil, err
	}

	return &auth, nil

}

// GetUserByEmail retrieves a user by email
func (ur *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User

	err := ur.DB.QueryRow(
		"SELECT user_id, username, email, COALESCE(provider, ''), COALESCE(provider_id, ''), COALESCE(provider_email, ''), created_at FROM users WHERE LOWER(email) = LOWER(?)",
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Provider, &user.ProviderID, &user.ProviderEmail, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil

}

// Authenticate validates a user's login credentials
func (ur *UserRepository) Authenticate(login models.UserLogin) (*models.User, error) {
	// Get the user by email
	user, err := ur.GetUserByEmail(login.Email)
	if err != nil {
		return nil, errors.New("email not found")
	}

	// Get the user's authentication data
	auth, err := ur.GetAuthByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	// Check the password
	if !utils.CheckPasswordHash(login.Password, auth.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (ur *UserRepository) GetCurrentUser(userID string) (*models.User, error) {

	var user models.User

	err := ur.DB.QueryRow(
		"SELECT user_id, username, email, COALESCE(provider, ''), COALESCE(provider_id, ''), COALESCE(provider_email, ''), created_at FROM users WHERE user_id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Provider, &user.ProviderID, &user.ProviderEmail, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil

}

// GetUserProfile retrieves complete user profile with statistics
func (ur *UserRepository) GetUserProfile(userID string) (*models.UserProfile, error) {
	// First get basic user info using existing method
	user, err := ur.GetCurrentUser(userID)
	if err != nil {
		return nil, err
	}

	// Get profile statistics
	stats, err := ur.GetProfileStats(userID)
	if err != nil {
		return nil, err
	}

	profile := &models.UserProfile{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		Stats:     *stats,
	}

	return profile, nil
}

// GetProfileStats calculates all profile statistics for a user
func (ur *UserRepository) GetProfileStats(userID string) (*models.ProfileStats, error) {
	stats := &models.ProfileStats{}

	// 1. Count total posts by user
	err := ur.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE user_id = ?", userID).Scan(&stats.TotalPosts)
	if err != nil {
		return nil, err
	}

	// 2. Count total comments by user
	err = ur.DB.QueryRow("SELECT COUNT(*) FROM comments WHERE user_id = ?", userID).Scan(&stats.TotalComments)
	if err != nil {
		return nil, err
	}

	// 3. Count posts this user has liked - FIXED: use post_reactions table
	err = ur.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM post_reactions 
		WHERE user_id = ? AND reaction_type = 1
	`, userID).Scan(&stats.PostsLiked)
	if err != nil {
		return nil, err
	}

	// 4. Count posts this user has commented on (NEW)
	err = ur.DB.QueryRow(`
		SELECT COUNT(DISTINCT p.post_id) 
		FROM posts p
		JOIN comments c ON p.post_id = c.post_id
		WHERE c.user_id = ?
	`, userID).Scan(&stats.PostsCommentedOn)
	if err != nil {
		return nil, err
	}

	// 5. Count total likes received (on posts + comments) - FIXED: use separate tables
	err = ur.DB.QueryRow(`
		SELECT (
			SELECT COALESCE(COUNT(*), 0) FROM post_reactions pr 
			JOIN posts p ON pr.post_id = p.post_id 
			WHERE pr.reaction_type = 1 AND p.user_id = ?
		) + (
			SELECT COALESCE(COUNT(*), 0) FROM comment_reactions cr 
			JOIN comments c ON cr.comment_id = c.comment_id 
			WHERE cr.reaction_type = 1 AND c.user_id = ?
		)
	`, userID, userID).Scan(&stats.LikesReceived)
	if err != nil {
		return nil, err
	}

	// 6. Count total dislikes received (on posts + comments) - FIXED: use separate tables
	err = ur.DB.QueryRow(`
		SELECT (
			SELECT COALESCE(COUNT(*), 0) FROM post_reactions pr 
			JOIN posts p ON pr.post_id = p.post_id 
			WHERE pr.reaction_type = 2 AND p.user_id = ?
		) + (
			SELECT COALESCE(COUNT(*), 0) FROM comment_reactions cr 
			JOIN comments c ON cr.comment_id = c.comment_id 
			WHERE cr.reaction_type = 2 AND c.user_id = ?
		)
	`, userID, userID).Scan(&stats.DislikesReceived)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// OAuthUserResult represents the result of creating or getting an OAuth user
type OAuthUserResult struct {
	User      *models.User
	IsNewUser bool
}

// CreateOrGetOAuthUser creates a new OAuth user or returns existing one
func (ur *UserRepository) CreateOrGetOAuthUser(googleUser *models.GoogleUserInfo) (*models.User, bool, error) {
	result, err := utils.ExecuteInTransactionWithResult(ur.DB, func(tx *sql.Tx) (OAuthUserResult, error) {
		// First, try to find existing user by OAuth provider and provider ID
		var existingUser models.User
		err := tx.QueryRow(`
			SELECT user_id, username, email, provider, provider_id, provider_email, created_at 
			FROM users 
			WHERE provider = 'google' AND provider_id = ?
		`, googleUser.ID).Scan(
			&existingUser.ID, &existingUser.Username, &existingUser.Email,
			&existingUser.Provider, &existingUser.ProviderID, &existingUser.ProviderEmail,
			&existingUser.CreatedAt,
		)
		
		if err == nil {
			// User exists, return it
			return OAuthUserResult{User: &existingUser, IsNewUser: false}, nil
		}
		
		if err != sql.ErrNoRows {
			// Database error
			return OAuthUserResult{}, err
		}
		
		// User doesn't exist, check if email is already taken by regular account
		var emailCount int
		err = tx.QueryRow("SELECT COUNT(*) FROM users WHERE LOWER(email) = LOWER(?) AND provider IS NULL", googleUser.Email).Scan(&emailCount)
		if err != nil {
			return OAuthUserResult{}, err
		}
		
		if emailCount > 0 {
			// Email exists for regular account, could link accounts but for now return error
			return OAuthUserResult{}, errors.New("email already registered with regular account, please use regular login")
		}
		
		// Generate username from email
		baseUsername := utils.GenerateUsernameFromEmail(googleUser.Email)
		username := baseUsername
		
		// Ensure username is unique
		counter := 1
		for {
			var usernameCount int
			err = tx.QueryRow("SELECT COUNT(*) FROM users WHERE LOWER(username) = LOWER(?)", username).Scan(&usernameCount)
			if err != nil {
				return OAuthUserResult{}, err
			}
			
			if usernameCount == 0 {
				break // Username is available
			}
			
			// Username taken, try with number suffix
			username = baseUsername + utils.GenerateUUIDToken()[:4]
			counter++
			
			if counter > 10 {
				// Fallback to UUID if we can't find unique username
				username = "user_" + utils.GenerateUUIDToken()[:8]
				break
			}
		}
		
		// Create new OAuth user
		userID := utils.GenerateUUIDToken()
		createdAt := time.Now()
		
		_, err = tx.Exec(`
			INSERT INTO users (user_id, username, email, provider, provider_id, provider_email, created_at) 
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, userID, username, googleUser.Email, "google", googleUser.ID, googleUser.Email, createdAt)
		
		if err != nil {
			return OAuthUserResult{}, err
		}
		
		// Return the new user
		newUser := &models.User{
			ID:            userID,
			Username:      username,
			Email:         googleUser.Email,
			Provider:      "google",
			ProviderID:    googleUser.ID,
			ProviderEmail: googleUser.Email,
			CreatedAt:     createdAt,
		}
		
		return OAuthUserResult{User: newUser, IsNewUser: true}, nil
	})
	
	if err != nil {
		return nil, false, err
	}
	
	return result.User, result.IsNewUser, nil
}

// GetOAuthUserByProvider gets OAuth user by provider and provider ID
func (ur *UserRepository) GetOAuthUserByProvider(provider, providerID string) (*models.User, error) {
	var user models.User
	
	err := ur.DB.QueryRow(`
		SELECT user_id, username, email, provider, provider_id, provider_email, created_at 
		FROM users 
		WHERE provider = ? AND provider_id = ?
	`, provider, providerID).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.Provider, &user.ProviderID, &user.ProviderEmail,
		&user.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("oauth user not found")
		}
		return nil, err
	}
	
	return &user, nil
}
