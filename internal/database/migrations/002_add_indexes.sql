-- Migration 002: Add Performance Indexes
-- Creates indexes for frequently queried columns and foreign keys

-- Index on contributions.derive_id for faster lookups of contributions by derive
CREATE INDEX IF NOT EXISTS idx_contributions_derive_id ON contributions(derive_id);

-- Index on contributions.created_at for sorting and latest contribution queries
CREATE INDEX IF NOT EXISTS idx_contributions_created_at ON contributions(created_at DESC);

-- Index on upload_logs.token_id for session tracking queries
CREATE INDEX IF NOT EXISTS idx_upload_logs_token_id ON upload_logs(token_id);

-- Index on upload_logs.session_number for session-specific queries
CREATE INDEX IF NOT EXISTS idx_upload_logs_session_number ON upload_logs(session_number);

-- Index on upload_logs.contribution_id for join performance
CREATE INDEX IF NOT EXISTS idx_upload_logs_contribution_id ON upload_logs(contribution_id);

-- Composite index for upload_logs queries filtering by token_id and session_number
CREATE INDEX IF NOT EXISTS idx_upload_logs_token_session ON upload_logs(token_id, session_number);

-- Index on upload_tokens.token for fast token validation
CREATE INDEX IF NOT EXISTS idx_upload_tokens_token ON upload_tokens(token);

-- Index on upload_tokens.is_active for filtering active tokens
CREATE INDEX IF NOT EXISTS idx_upload_tokens_is_active ON upload_tokens(is_active);

-- Index on bag_requests.handled for filtering by status
CREATE INDEX IF NOT EXISTS idx_bag_requests_handled ON bag_requests(handled);

-- Index on bag_requests.created_at for sorting
CREATE INDEX IF NOT EXISTS idx_bag_requests_created_at ON bag_requests(created_at DESC);

-- Index on deriven.number for lookups by derive number
CREATE INDEX IF NOT EXISTS idx_deriven_number ON deriven(number);
