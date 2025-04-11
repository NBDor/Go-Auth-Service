-- Migration: 001_initial_schema (rollback)

-- Drop all tables
DROP TABLE IF EXISTS revoked_tokens;
DROP TABLE IF EXISTS user_metadata;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS schema_migrations;
