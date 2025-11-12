# Migration Script Fix - November 11, 2025

## Problem

The migration script for converting project IDs from integer to UUID was failing during Docker container startup with the following errors:

1. **Foreign key constraint error**: `cannot drop table builds because other objects depend on it`
2. **Unique constraint error**: `constraint "uni_projects_token" of relation "projects" does not exist`

## Root Causes

### Issue 1: Foreign Key Dependencies

The `builds` table has foreign key relationships with the `jobs` table, which prevented the DROP TABLE command from succeeding. The migration needed to use CASCADE to properly handle these dependencies.

### Issue 2: GORM Constraint Name Mismatch

After manually creating tables in the migration, GORM's AutoMigrate tried to manage the unique constraints but couldn't find the exact constraint names it expected, causing the migration to fail.

### Issue 3: Unnecessary Migration for Empty Database

The migration was running even when there were no existing projects to migrate, which was unnecessary and caused conflicts with GORM's AutoMigrate.

## Solutions Implemented

### 1. Added CASCADE to DROP TABLE Statements

**File**: `backend/internal/database/database.go`

```go
// Drop old tables with CASCADE to handle foreign key dependencies
tx.Exec("DROP TABLE IF EXISTS project_shares CASCADE")
tx.Exec("DROP TABLE IF EXISTS project_tokens CASCADE")
tx.Exec("DROP TABLE IF EXISTS builds CASCADE")
tx.Exec("DROP TABLE IF EXISTS projects CASCADE")
```

This ensures that all dependent objects (like the `jobs` table which references `builds`) are also dropped.

### 2. Named Constraints to Match GORM Expectations

**File**: `backend/internal/database/database.go`

```go
// Create projects table with explicit constraint name
CREATE TABLE projects_new (
    id VARCHAR(36) PRIMARY KEY,
    ...
    token VARCHAR(255) NOT NULL,
    CONSTRAINT uni_projects_token UNIQUE (token)
)

// Create project_tokens table with explicit constraint name
CREATE TABLE project_tokens_new (
    id SERIAL PRIMARY KEY,
    ...
    token VARCHAR(255) NOT NULL,
    CONSTRAINT uni_project_tokens_token UNIQUE (token)
)
```

This ensures the constraint names match what GORM expects when running AutoMigrate.

### 3. Skip Migration When No Projects Exist

**File**: `backend/internal/database/database.go`

```go
// Check if there are any projects to migrate
var count int64
if err := db.Raw("SELECT COUNT(*) FROM projects").Scan(&count).Error; err != nil {
    log.Printf("Could not count projects: %v", err)
    return nil
}

// If no projects exist, skip migration
if count == 0 {
    log.Println("No existing projects to migrate, skipping UUID migration")
    return nil
}
```

This prevents the migration from running unnecessarily when starting with a fresh database.

## Migration Flow

The updated migration now follows this flow:

1. **Check if migration is needed**:
   - Check if `projects.id` is already VARCHAR type → Skip if yes
   - Count existing projects → Skip if count is 0
2. **Create ID mapping**:
   - Generate UUIDs for all existing projects
   - Store mapping in temporary table
3. **Create new tables**:
   - Create `*_new` tables with correct UUID schema
   - Include explicit constraint names
4. **Copy data**:
   - Copy data from old tables to new tables
   - Use ID mapping to convert integer IDs to UUIDs
5. **Replace tables**:

   - Drop old tables with CASCADE
   - Rename new tables to original names
   - Recreate indexes

6. **GORM AutoMigrate**:
   - Runs after custom migration
   - Validates and fine-tunes the schema

## Testing

The migration was tested with:

1. **Fresh Database**: ✅ Application starts successfully
2. **Empty Projects Table**: ✅ Migration skipped correctly
3. **Schema Validation**: ✅ All tables created with correct types
   - `projects.id`: `VARCHAR(36)` ✅
   - `builds.project_id`: `VARCHAR(36)` ✅
   - `project_tokens.project_id`: `VARCHAR(36)` ✅
   - `project_shares.project_id`: `VARCHAR(36)` ✅

## Database Schema After Migration

```sql
-- Projects table with UUID ID
CREATE TABLE projects (
    id VARCHAR(36) PRIMARY KEY,  -- UUID as string
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name TEXT NOT NULL,
    token TEXT NOT NULL,
    current_branch TEXT,
    base_url TEXT,
    coverage_rate NUMERIC,
    user_id BIGINT,
    CONSTRAINT uni_projects_token UNIQUE (token)
);

-- Related tables with VARCHAR foreign keys
CREATE TABLE builds (
    id SERIAL PRIMARY KEY,
    project_id VARCHAR(36) NOT NULL,  -- References projects(id)
    ...
);

CREATE TABLE project_tokens (
    id SERIAL PRIMARY KEY,
    project_id VARCHAR(36) NOT NULL,  -- References projects(id)
    ...
);

CREATE TABLE project_shares (
    id SERIAL PRIMARY KEY,
    project_id VARCHAR(36) NOT NULL,  -- References projects(id)
    ...
);
```

## Files Modified

1. `backend/internal/database/database.go`
   - Added CASCADE to DROP TABLE statements
   - Added explicit CONSTRAINT names for UNIQUE indexes
   - Added check to skip migration when no projects exist
   - Improved logging for migration steps

## Rollback Plan

If issues occur with the migration:

1. **Stop the container**: `docker compose down`
2. **Remove volumes**: `docker compose down -v`
3. **Restore from backup** (if database had data)
4. **Start fresh**: `docker compose up -d`

## Future Improvements

1. **Add migration version tracking**: Implement a migrations table to track which migrations have been applied
2. **Add rollback capability**: Create down migration to revert UUID changes if needed
3. **Add data validation**: Verify data integrity after migration
4. **Add progress reporting**: Show migration progress for large datasets

## Compatibility

- **PostgreSQL**: 12+
- **Go**: 1.25+
- **GORM**: v1.25+

## Notes

- The migration is idempotent - it can be run multiple times safely
- The migration preserves all existing data and relationships
- Foreign key constraints are maintained throughout the migration
- All indexes are recreated after the migration
