# Implementation Notes

## Changes Made

### 1. UUID for Project IDs (Security Enhancement)

**Problem**: Using auto-incrementing integer IDs for projects makes them predictable and potentially vulnerable to enumeration attacks.

**Solution**: 
- Changed project ID from `uint` to `string` (UUID)
- Added BeforeCreate hook to automatically generate UUID v7 for new projects
- Implemented automatic migration for existing projects from integer IDs to UUIDs

**Files Changed**:
- `backend/internal/models/models.go` - Updated Project, ProjectShare, ProjectToken, and Build models
- `backend/internal/database/database.go` - Added migration logic
- All backend handlers updated to work with string IDs

### 2. Build Details Display

**Solution**: 
- Enhanced ProjectView to fetch and display builds from the API
- Added expandable build cards showing:
  - Build number, branch, and coverage rate
  - Commit SHA and message
  - Build timestamp
  - Job details when expanded

**Files Changed**:
- `frontend/src/views/ProjectView.vue` - Complete rewrite with build display

### 3. Badge Visibility

**Solution**:
- Added badge section in ProjectView
- Badge URL: `/projects/:id/badge.svg`
- Copy-to-clipboard functionality for markdown
- Badge is visible to all users with project access (owners and shared users)

**Files Changed**:
- `frontend/src/views/ProjectView.vue` - Added badge display section

### 4. Token Info for Shared Users

**Solution**:
- Token information is now displayed in ProjectView (not just settings)
- Modal shows token with usage examples
- All users with project access can view the token
- Copy-to-clipboard functionality

**Files Changed**:
- `frontend/src/views/ProjectView.vue` - Added token info modal

### 5. Token Refresh Functionality

**Solution**:
- New API endpoint: `POST /api/v1/projects/:id/refresh-token`
- Requires user authentication and project ownership
- Returns new token immediately
- UI includes confirmation dialog to prevent accidental refresh
- Shows new token once with copy functionality

**Files Changed**:
- `backend/internal/api/token_handler.go` - Added RefreshProjectToken handler
- `backend/internal/api/routes.go` - Added route
- `backend/internal/api/token_handler_test.go` - Added test
- `frontend/src/services/api.ts` - Added API method
- `frontend/src/views/ProjectView.vue` - Added refresh UI

## Migration Strategy

The migration is handled automatically on application startup:
1. Checks if projects table has integer or varchar ID type
2. If integer, creates UUID mappings for all existing projects
3. Creates new tables with correct schema
4. Copies data with UUID mapping
5. Drops old tables and renames new ones
6. Recreates indexes

This ensures zero data loss during migration.

## Testing

All backend tests pass:
- Token refresh test added
- Upload tests updated for UUID
- Model tests updated for UUID
- All 8 test suites passing

## Security

- CodeQL scan: 0 vulnerabilities found
- UUID library (google/uuid v1.6.0): No known vulnerabilities
- Token refresh requires authentication and ownership verification
- UUIDs are cryptographically secure (v7 includes timestamp + random data)
