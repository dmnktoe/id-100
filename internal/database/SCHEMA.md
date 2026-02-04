# Database Schema Overview

This document provides a comprehensive overview of the database schema for the id-100 application.

## Schema Evolution

The database schema is managed through versioned migration files located in `internal/database/migrations/`.

### Migration History

| Version | File | Description |
|---------|------|-------------|
| 000 | `000_schema_migrations.sql` | Migration tracking table |
| 001 | `001_initial_schema.sql` | Core tables and relationships |
| 002 | `002_add_indexes.sql` | Performance indexes |

## Database Tables

### 1. deriven
Main derives/tasks that users can contribute to.

**Columns:**
- `id` (SERIAL PRIMARY KEY) - Unique identifier
- `number` (INTEGER NOT NULL UNIQUE) - Derive number (user-facing)
- `title` (TEXT NOT NULL) - Title of the derive
- `description` (TEXT) - Detailed description
- `points` (INTEGER DEFAULT 0) - Points assigned to derive (for badges)
- `created_at` (TIMESTAMPTZ DEFAULT NOW()) - Creation timestamp

**Indexes:**
- Primary key on `id`
- Unique index on `number`
- Index on `number` (idx_deriven_number)

**Relationships:**
- One-to-many with `contributions`

---

### 2. contributions
User contributions (images) for derives.

**Columns:**
- `id` (SERIAL PRIMARY KEY) - Unique identifier
- `derive_id` (INTEGER NOT NULL) - Foreign key to deriven.id
- `image_url` (TEXT NOT NULL) - URL/path to the uploaded image
- `image_lqip` (TEXT) - Low-quality image placeholder data
- `user_name` (TEXT DEFAULT '') - Name of contributor
- `user_city` (TEXT DEFAULT '') - City of contributor
- `created_at` (TIMESTAMPTZ DEFAULT NOW()) - Contribution timestamp

**Indexes:**
- Primary key on `id`
- Index on `derive_id` (idx_contributions_derive_id)
- Index on `created_at` DESC (idx_contributions_created_at)

**Relationships:**
- Many-to-one with `deriven` (via derive_id)
- One-to-many with `upload_logs`

**Constraints:**
- Foreign key: `derive_id` REFERENCES `deriven(id)` ON DELETE CASCADE

---

### 3. upload_tokens
Authorization tokens for uploads with session tracking.

**Columns:**
- `id` (SERIAL PRIMARY KEY) - Unique identifier
- `token` (TEXT NOT NULL UNIQUE) - Authorization token string
- `bag_name` (TEXT) - Name/identifier for the bag
- `current_player` (TEXT) - Current player using this token
- `current_player_city` (TEXT DEFAULT '') - City of current player
- `is_active` (BOOLEAN DEFAULT TRUE) - Whether token is active
- `max_uploads` (INTEGER DEFAULT 10) - Maximum uploads per session
- `total_uploads` (INTEGER DEFAULT 0) - Total uploads across all sessions
- `total_sessions` (INTEGER DEFAULT 1) - Number of sessions completed
- `session_started_at` (TIMESTAMPTZ) - When current session started
- `created_at` (TIMESTAMPTZ DEFAULT NOW()) - Token creation timestamp

**Indexes:**
- Primary key on `id`
- Unique index on `token`
- Index on `token` (idx_upload_tokens_token)
- Index on `is_active` (idx_upload_tokens_is_active)

**Relationships:**
- One-to-many with `upload_logs`

---

### 4. upload_logs
Audit trail of all uploads with session information.

**Columns:**
- `id` (SERIAL PRIMARY KEY) - Unique identifier
- `token_id` (INTEGER NOT NULL) - Foreign key to upload_tokens.id
- `contribution_id` (INTEGER NOT NULL) - Foreign key to contributions.id
- `derive_number` (INTEGER NOT NULL) - Number of derive contributed to
- `player_name` (TEXT) - Name of player who uploaded
- `session_number` (INTEGER NOT NULL) - Session number for this token
- `uploaded_at` (TIMESTAMPTZ DEFAULT NOW()) - Upload timestamp

**Indexes:**
- Primary key on `id`
- Index on `token_id` (idx_upload_logs_token_id)
- Index on `session_number` (idx_upload_logs_session_number)
- Index on `contribution_id` (idx_upload_logs_contribution_id)
- Composite index on `(token_id, session_number)` (idx_upload_logs_token_session)

**Relationships:**
- Many-to-one with `upload_tokens` (via token_id)
- Many-to-one with `contributions` (via contribution_id)

**Constraints:**
- Foreign key: `token_id` REFERENCES `upload_tokens(id)` ON DELETE CASCADE
- Foreign key: `contribution_id` REFERENCES `contributions(id)` ON DELETE CASCADE

---

### 5. bag_requests
User requests for new bags/tokens.

**Columns:**
- `id` (SERIAL PRIMARY KEY) - Unique identifier
- `email` (TEXT NOT NULL) - Email address of requester
- `created_at` (TIMESTAMPTZ DEFAULT NOW()) - Request timestamp
- `handled` (BOOLEAN DEFAULT FALSE) - Whether request has been processed

**Indexes:**
- Primary key on `id`
- Index on `handled` (idx_bag_requests_handled)
- Index on `created_at` DESC (idx_bag_requests_created_at)

**Relationships:**
- None (standalone table)

---

### 6. schema_migrations
Tracks which database migrations have been applied.

**Columns:**
- `version` (INTEGER PRIMARY KEY) - Migration version number
- `description` (TEXT NOT NULL) - Human-readable description
- `applied_at` (TIMESTAMPTZ DEFAULT NOW()) - When migration was applied

**Relationships:**
- None (internal tracking table)

---

## Entity Relationship Diagram

```
deriven (1) ----< (many) contributions (1) ----< (many) upload_logs
                                                           ^
                                                           |
                                                         (many)
                                                           |
upload_tokens (1) ----------------------------------------<

bag_requests (standalone)

schema_migrations (internal)
```

## Key Design Decisions

1. **Cascade Deletes**: Foreign keys use `ON DELETE CASCADE` to maintain referential integrity
2. **Default Values**: Sensible defaults for boolean and text fields to avoid NULL issues
3. **Timestamps**: All tables have `created_at` for audit trailing
4. **Indexes**: Strategic indexes on foreign keys and frequently queried columns
5. **Idempotency**: All migrations use `IF NOT EXISTS` for safe re-running

## Query Patterns

Common query patterns and their optimized indexes:

1. **Get contributions by derive**: Uses `idx_contributions_derive_id`
2. **List recent contributions**: Uses `idx_contributions_created_at`
3. **Validate token**: Uses `idx_upload_tokens_token`
4. **Get active tokens**: Uses `idx_upload_tokens_is_active`
5. **Track session uploads**: Uses `idx_upload_logs_token_session`
6. **Filter bag requests**: Uses `idx_bag_requests_handled`

## Data Types

- **SERIAL**: Auto-incrementing integer (PostgreSQL-specific)
- **INTEGER**: Standard 4-byte integer
- **TEXT**: Variable-length character string
- **BOOLEAN**: True/false values
- **TIMESTAMPTZ**: Timestamp with timezone (recommended for PostgreSQL)
