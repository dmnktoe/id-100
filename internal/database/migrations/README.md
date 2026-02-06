# Database Migrations

This directory contains SQL migration files for the ID-100 application.

## Overview

Migrations are automatically executed when the application starts. The system tracks which migrations have been applied using a `schema_migrations` table.

## Migration Files

Migrations are named with the pattern: `NNN_description.sql` where:
- `NNN` is a zero-padded version number (e.g., 001, 002, 003)
- `description` is a brief description of what the migration does

### Current Migrations

- **001_create_initial_tables.sql**: Creates the core database schema including:
  - `deriven`: Main derive/challenge definitions
  - `contributions`: User contributions/uploads
  - `upload_tokens`: Upload token management
  - `upload_logs`: Upload audit logs
  - `bag_requests`: Bag/tool requests from users

- **002_insert_initial_deriven.sql**: Inserts initial deriven (challenge) data
  - Populated by converting a Supabase export using `scripts/convert-deriven-export.sh`
  - A placeholder `deriven_rows.sql` file exists in this directory
  - Replace the placeholder with your Supabase export and run the conversion script
  - See `docs/ADDING_DERIVEN_DATA.md` for detailed instructions

## Adding New Migrations

1. Create a new `.sql` file in this directory
2. Name it with the next sequential number: `00X_your_description.sql`
3. Write idempotent SQL (use `IF NOT EXISTS` where appropriate)
4. The migration will run automatically on next application start

## Migration System

The migration system:
- Reads migration files from this directory (embedded at compile time)
- Tracks applied migrations in the `schema_migrations` table
- Runs migrations in order by version number
- Wraps each migration in a transaction
- Logs success/failure for each migration

## Example Migration

```sql
-- Migration: 002_add_user_preferences.sql
-- Description: Adds user preferences table
-- Date: 2024-02-06

CREATE TABLE IF NOT EXISTS user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    preference_key TEXT NOT NULL,
    preference_value TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, preference_key)
);

CREATE INDEX IF NOT EXISTS idx_user_preferences_user_id 
    ON user_preferences(user_id);
```

## Rollback

The system does not support automatic rollbacks. To revert a migration:
1. Create a new migration that undoes the changes
2. Never delete or modify existing migration files
