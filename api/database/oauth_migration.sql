-- OAuth Migration Script for Users Table
-- Add OAuth support to existing users table

-- Add OAuth columns to users table
ALTER TABLE users ADD COLUMN provider VARCHAR(50) DEFAULT NULL;
ALTER TABLE users ADD COLUMN provider_id VARCHAR(255) DEFAULT NULL;
ALTER TABLE users ADD COLUMN provider_email VARCHAR(255) DEFAULT NULL;

-- Create indexes for OAuth lookups
CREATE INDEX IF NOT EXISTS idx_users_provider_id ON users(provider, provider_id);
CREATE INDEX IF NOT EXISTS idx_users_provider_email ON users(provider, provider_email);

-- Make password_hash optional for OAuth users
-- Note: This requires updating the table to allow NULL passwords for OAuth users
-- In the application logic, we check if provider is set before requiring password