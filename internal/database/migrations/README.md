# Database Migrations

This directory contains SQL migration files for the id-100 application database.

## Migration File Structure

Migrations are organized as numbered SQL files in a modular, structured manner:

```
migrations/
├── 000_schema_migrations.sql  # Migration tracking table (auto-applied)
├── 001_initial_schema.sql     # Core table schemas
├── 002_add_indexes.sql        # Performance indexes
```

## Naming Convention

Migration files follow the pattern: `{version}_{description}.sql`

- **Version**: 3-digit number (e.g., 001, 002, 003)
- **Description**: Snake_case description of what the migration does

## How Migrations Work

1. **Automatic Execution**: Migrations run automatically when the application starts via `database.Init()`
2. **Version Tracking**: The `schema_migrations` table tracks which migrations have been applied
3. **Idempotent**: Migrations use `IF NOT EXISTS` clauses to be safely re-runnable
4. **Ordered**: Migrations are executed in numerical order by version number
5. **Embedded**: Migration files are embedded in the binary using Go's `embed` directive

## Current Migrations

### 000_schema_migrations.sql
Creates the `schema_migrations` table to track applied migrations.

### 001_initial_schema.sql
Creates all core tables:
- **deriven**: Main derives/tasks that users can contribute to
- **contributions**: User image contributions for derives
- **upload_tokens**: Authorization tokens for uploads with session tracking
- **upload_logs**: Audit trail of all uploads
- **bag_requests**: User requests for new bags/tokens

### 002_add_indexes.sql
Adds performance indexes on:
- Foreign key columns for faster joins
- Frequently filtered columns (is_active, handled)
- Sorting columns (created_at, uploaded_at)
- Composite indexes for common query patterns

## Adding New Migrations

To add a new migration:

1. Create a new SQL file with the next version number:
   ```bash
   touch internal/database/migrations/003_add_new_feature.sql
   ```

2. Write your SQL statements using `IF NOT EXISTS` for idempotency:
   ```sql
   -- Migration 003: Add New Feature
   -- Description of what this migration does
   
   ALTER TABLE table_name ADD COLUMN IF NOT EXISTS new_column TEXT;
   CREATE INDEX IF NOT EXISTS idx_name ON table_name(column);
   ```

3. The migration will automatically run on next application start

## Database Schema

### Tables

#### deriven
Stores the main derives/tasks.
- Primary key: `id`
- Unique key: `number`

#### contributions
Stores user contributions (images).
- Primary key: `id`
- Foreign key: `derive_id` → `deriven(id)`

#### upload_tokens
Manages authorization tokens.
- Primary key: `id`
- Unique key: `token`

#### upload_logs
Audit trail of uploads.
- Primary key: `id`
- Foreign keys: `token_id` → `upload_tokens(id)`, `contribution_id` → `contributions(id)`

#### bag_requests
User requests for tokens.
- Primary key: `id`

## Rollback Strategy

Currently, the migration system does not support automatic rollbacks. To rollback:

1. Manually write and execute rollback SQL
2. Delete the version from `schema_migrations` table
3. Restart the application to re-apply if needed

## Local Development

For local development with a fresh database:

```bash
# Start local PostgreSQL
make docker-db

# Migrations run automatically when you start the app
make run
```

## Production Deployment

Migrations run automatically on application startup. Ensure:
1. Database backups are in place before deploying
2. Migrations are tested in staging environment
3. Monitor application logs during deployment for migration errors
