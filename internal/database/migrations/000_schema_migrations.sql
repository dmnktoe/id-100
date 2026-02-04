-- Migration 003: Add Schema Version Tracking
-- Creates a table to track which migrations have been applied

CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    description TEXT NOT NULL,
    applied_at TIMESTAMPTZ DEFAULT NOW()
);
