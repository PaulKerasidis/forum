package database

// TableCreationStatements contains all SQL statements for optimized table creation
var TableCreationStatements = []string{
	// Users table - updated with OAuth support
	`CREATE TABLE IF NOT EXISTS users (
		user_id TEXT PRIMARY KEY NOT NULL UNIQUE,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT, -- Made optional for OAuth users
		provider TEXT, -- OAuth provider (google, github, etc.)
		provider_id TEXT, -- User ID from OAuth provider
		provider_email TEXT, -- Email from OAuth provider
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`,

	// Sessions table - unchanged, already optimized
	`CREATE TABLE IF NOT EXISTS sessions (
		user_id TEXT PRIMARY KEY NOT NULL UNIQUE,
		session_id TEXT NOT NULL UNIQUE,
		ip_address TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`,

	// Categories table - unchanged, already optimized
	`CREATE TABLE IF NOT EXISTS categories (
		category_id TEXT PRIMARY KEY NOT NULL UNIQUE,
		category_name TEXT NOT NULL UNIQUE
	);`,

	// Posts table - CLEAN, no denormalized counts
	`CREATE TABLE IF NOT EXISTS posts (
		post_id TEXT PRIMARY KEY NOT NULL UNIQUE,
		user_id TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NULL,


		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`,

	// Post categories junction table - unchanged
	`CREATE TABLE IF NOT EXISTS post_categories (
		post_id TEXT NOT NULL,
		category_id TEXT NOT NULL,
		PRIMARY KEY (post_id, category_id),
		FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
		FOREIGN KEY (category_id) REFERENCES categories(category_id) ON DELETE CASCADE
	);`,

	// Comments table - CLEAN, no denormalized counts
	`CREATE TABLE IF NOT EXISTS comments (
		comment_id TEXT PRIMARY KEY NOT NULL UNIQUE,
		post_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NULL,
		
		FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);`,

	// Post reactions table - SEPARATED from comments for better performance
	`CREATE TABLE IF NOT EXISTS post_reactions (
		user_id TEXT NOT NULL,
		post_id TEXT NOT NULL,
		reaction_type INTEGER NOT NULL, -- 1 for like, 2 for dislike
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		
		-- Natural primary key - no UUID needed, prevents duplicate reactions
		PRIMARY KEY (user_id, post_id),
		
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
		FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
		
		-- Ensure valid reaction types
		CHECK (reaction_type IN (1, 2))
	);`,

	// Comment reactions table - SEPARATED for better performance
	`CREATE TABLE IF NOT EXISTS comment_reactions (
		user_id TEXT NOT NULL,
		comment_id TEXT NOT NULL,
		reaction_type INTEGER NOT NULL, -- 1 for like, 2 for dislike
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		
		-- Natural primary key - no UUID needed, prevents duplicate reactions
		PRIMARY KEY (user_id, comment_id),
		
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
		FOREIGN KEY (comment_id) REFERENCES comments(comment_id) ON DELETE CASCADE,
		
		-- Ensure valid reaction types
		CHECK (reaction_type IN (1, 2))
	);`,
}

// IndexCreationStatements contains ESSENTIAL indexes only - what you'll actually use
var IndexCreationStatements = []string{
	// Authentication indexes (used every request)
	`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,                 // Login by email
	`CREATE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);`, // Session validation
	
	// OAuth indexes
	`CREATE INDEX IF NOT EXISTS idx_users_provider_id ON users(provider, provider_id);`, // OAuth user lookup
	`CREATE INDEX IF NOT EXISTS idx_users_provider_email ON users(provider, provider_email);`, // OAuth email lookup

	// Core post browsing indexes (main forum functionality)
	`CREATE INDEX IF NOT EXISTS idx_posts_created_desc ON posts(created_at DESC);`,                           // Homepage post list
	`CREATE INDEX IF NOT EXISTS idx_post_categories_category_post ON post_categories(category_id, post_id);`, // Posts by category

	// Comment indexes (viewing posts with comments)
	`CREATE INDEX IF NOT EXISTS idx_comments_post_created ON comments(post_id, created_at ASC);`, // Comments for a post

	// Reaction indexes (like/dislike counts)
	`CREATE INDEX IF NOT EXISTS idx_post_reactions_post_type ON post_reactions(post_id, reaction_type);`,             // Post reaction counts
	`CREATE INDEX IF NOT EXISTS idx_comment_reactions_comment_type ON comment_reactions(comment_id, reaction_type);`, // Comment reaction counts
}

// // WALModeStatements contains SQL statements for enabling WAL mode and performance optimization
// var WALModeStatements = []string{
// 	// Enable WAL mode for better concurrency
// 	`PRAGMA journal_mode=WAL;`,

// 	// Optimize WAL performance
// 	`PRAGMA synchronous=NORMAL;`,  // Faster than FULL, still safe
// 	`PRAGMA cache_size=10000;`,    // Larger cache for better performance
// 	`PRAGMA temp_store=memory;`,   // Keep temporary data in memory
// 	`PRAGMA mmap_size=268435456;`, // 256MB memory mapping for larger databases

// 	// WAL checkpoint optimization
// 	`PRAGMA wal_autocheckpoint=1000;`, // Checkpoint every 1000 pages
// }
