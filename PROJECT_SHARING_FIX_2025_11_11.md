# Project Sharing UUID Query Fix - November 11, 2025

## Problem

The project sharing functionality was failing with SQL errors when trying to fetch projects by UUID. The errors were:

1. **"column 'nan' does not exist"** - GORM was trying to interpret UUID strings as numbers
2. **"trailing junk after numeric literal"** - PostgreSQL couldn't parse UUID strings as numbers

## Root Cause

After migrating project IDs from integers to UUID strings, the code was still using GORM's `query.First(&project, projectID)` method, which attempts to use the `projectID` parameter as a primary key value. When the primary key is a string (UUID), GORM generates SQL like:

```sql
SELECT * FROM "projects" WHERE 87effc2b-96d8-4da8-8775-011731278d9f AND "projects"."deleted_at" IS NULL
```

PostgreSQL interprets this as trying to use the UUID string as a numeric WHERE condition, causing the parsing errors.

## Solution

Changed all instances of `query.First(&project, projectID)` to `query.Where("id = ?", projectID).First(&project)` to explicitly specify the WHERE clause for string UUIDs.

## Files Modified

### backend/internal/api/project_handler.go

- `Get` function: Fixed project lookup
- `Update` function: Fixed project lookup
- `Delete` function: Fixed project lookup
- `GetShares` function: Fixed project lookup
- `CreateShare` function: Fixed project lookup
- `DeleteShare` function: Fixed project lookup

### backend/internal/api/handlers.go

- `BadgeHandler.GetBadge` function: Fixed project lookup

### backend/internal/database/database.go

- Removed unused `uuid` import that was causing compilation errors

## Code Changes

**Before:**

```go
if err := query.First(&project, projectID).Error; err != nil {
```

**After:**

```go
if err := query.Where("id = ?", projectID).First(&project).Error; err != nil {
```

## Affected Functions

1. **ProjectHandler.Get** - Get single project details
2. **ProjectHandler.Update** - Update project information
3. **ProjectHandler.Delete** - Delete project
4. **ProjectHandler.GetShares** - List project group shares
5. **ProjectHandler.CreateShare** - Create new group share
6. **ProjectHandler.DeleteShare** - Remove group share
7. **BadgeHandler.GetBadge** - Generate coverage badge

## Testing

The fix was verified by:

1. ✅ Application compiles successfully
2. ✅ Application starts without errors
3. ✅ No more SQL parsing errors in logs
4. ✅ Project sharing endpoints are accessible

## Database Schema

The project table now uses UUID strings:

```sql
projects (
    id VARCHAR(36) PRIMARY KEY,  -- UUID as string
    name TEXT NOT NULL,
    token TEXT NOT NULL,
    -- ... other fields
)
```

## Compatibility

- **GORM**: Works with string primary keys when using explicit WHERE clauses
- **PostgreSQL**: Handles VARCHAR UUID comparisons correctly
- **API**: Maintains backward compatibility with UUID string parameters

## Notes

This fix ensures that all project-related operations work correctly with the new UUID-based project IDs. The explicit `Where("id = ?", projectID)` pattern should be used for all string primary key lookups in GORM to avoid similar issues in the future.</content>
<parameter name="filePath">/home/coder/Librecov/PROJECT_SHARING_FIX_2025_11_11.md
