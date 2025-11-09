# API Documentation

LibreCov provides a RESTful API that is mostly compatible with the Coveralls API.

## Base URL

```
http://localhost:4000/api/v1
```

## Authentication

Most endpoints require authentication. Include the token in the request:

### Header (Recommended)
```
Authorization: Bearer YOUR_TOKEN
```

### Query Parameter
```
GET /api/v1/projects?token=YOUR_TOKEN
```

## Endpoints

### Authentication

#### `GET /auth/login`
Initiates OIDC login flow. Redirects to OIDC provider.

**Response**: Redirect to OIDC provider

---

#### `GET /auth/callback`
Handles OIDC callback after authentication.

**Query Parameters:**
- `code` - Authorization code from OIDC provider
- `state` - State parameter for CSRF protection

**Response:**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "admin": false
  },
  "token": "your-access-token"
}
```

---

#### `POST /auth/logout`
Logs out the current user.

**Headers:**
- `Authorization: Bearer TOKEN`

**Response:**
```json
{
  "message": "Logged out"
}
```

---

#### `GET /auth/me`
Get current user information.

**Headers:**
- `Authorization: Bearer TOKEN`

**Response:**
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "admin": false,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

### Projects

#### `GET /api/v1/projects`
List all projects for the authenticated user.

**Headers:**
- `Authorization: Bearer TOKEN`

**Response:**
```json
[
  {
    "id": 1,
    "name": "My Project",
    "token": "project-token-123",
    "current_branch": "main",
    "base_url": "https://github.com/user/repo",
    "coverage_rate": 85.5,
    "user_id": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

---

#### `POST /api/v1/projects`
Create a new project.

**Headers:**
- `Authorization: Bearer TOKEN`
- `Content-Type: application/json`

**Request Body:**
```json
{
  "name": "My Project",
  "current_branch": "main",
  "base_url": "https://github.com/user/repo"
}
```

**Response:**
```json
{
  "id": 1,
  "name": "My Project",
  "token": "project-token-123",
  "current_branch": "main",
  "base_url": "https://github.com/user/repo",
  "coverage_rate": 0,
  "user_id": 1,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

#### `GET /api/v1/projects/:id`
Get a specific project.

**Headers:**
- `Authorization: Bearer TOKEN`

**Response:**
```json
{
  "id": 1,
  "name": "My Project",
  "token": "project-token-123",
  "current_branch": "main",
  "base_url": "https://github.com/user/repo",
  "coverage_rate": 85.5,
  "user_id": 1,
  "builds": [...],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

#### `PUT /api/v1/projects/:id`
Update a project.

**Headers:**
- `Authorization: Bearer TOKEN`
- `Content-Type: application/json`

**Request Body:**
```json
{
  "name": "Updated Project Name",
  "current_branch": "develop",
  "base_url": "https://github.com/user/new-repo"
}
```

**Response:** Updated project object

---

#### `DELETE /api/v1/projects/:id`
Delete a project.

**Headers:**
- `Authorization: Bearer TOKEN`

**Response:**
```json
{
  "message": "Project deleted"
}
```

---

### Builds

#### `GET /api/v1/projects/:projectId/builds`
List all builds for a project.

**Headers:**
- `Authorization: Bearer TOKEN`

**Response:**
```json
[
  {
    "id": 1,
    "project_id": 1,
    "build_num": 42,
    "branch": "main",
    "commit_sha": "abc123def456",
    "commit_msg": "Fix bug",
    "coverage_rate": 87.5,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

---

#### `GET /api/v1/builds/:id`
Get a specific build.

**Headers:**
- `Authorization: Bearer TOKEN`

**Response:**
```json
{
  "id": 1,
  "project_id": 1,
  "build_num": 42,
  "branch": "main",
  "commit_sha": "abc123def456",
  "commit_msg": "Fix bug",
  "coverage_rate": 87.5,
  "jobs": [...],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

### Jobs

#### `GET /api/v1/builds/:buildId/jobs`
List all jobs for a build.

**Headers:**
- `Authorization: Bearer TOKEN`

**Response:**
```json
[
  {
    "id": 1,
    "build_id": 1,
    "job_number": "1.1",
    "coverage_rate": 88.5,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

---

#### `GET /api/v1/jobs/:id`
Get a specific job.

**Headers:**
- `Authorization: Bearer TOKEN`

**Response:**
```json
{
  "id": 1,
  "build_id": 1,
  "job_number": "1.1",
  "coverage_rate": 88.5,
  "data": "{...}",
  "files": [...],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

#### `POST /api/v1/jobs`
Create a new job (coverage data).

**Headers:**
- `Authorization: Bearer TOKEN` or project token
- `Content-Type: application/json`

**Request Body:** Coveralls JSON format

**Response:**
```json
{
  "message": "Not implemented yet"
}
```

---

### Files

#### `GET /api/v1/jobs/:jobId/files`
List all files for a job.

**Headers:**
- `Authorization: Bearer TOKEN`

**Response:**
```json
[
  {
    "id": 1,
    "job_id": 1,
    "name": "src/main.go",
    "coverage": "[1, 2, 0, 1, null]",
    "coverage_rate": 75.0,
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

---

#### `GET /api/v1/files/:id`
Get a specific file with source and coverage.

**Headers:**
- `Authorization: Bearer TOKEN`

**Response:**
```json
{
  "id": 1,
  "job_id": 1,
  "name": "src/main.go",
  "coverage": "[1, 2, 0, 1, null]",
  "source": "package main\n\nfunc main() {\n  ...\n}",
  "coverage_rate": 75.0,
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

### Admin (Admin Only)

#### `GET /api/v1/admin/users`
List all users.

**Headers:**
- `Authorization: Bearer ADMIN_TOKEN`

**Response:**
```json
[
  {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "admin": false,
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

---

#### `GET /api/v1/admin/users/:id`
Get a specific user.

**Headers:**
- `Authorization: Bearer ADMIN_TOKEN`

**Response:** User object with projects

---

#### `PUT /api/v1/admin/users/:id`
Update a user.

**Headers:**
- `Authorization: Bearer ADMIN_TOKEN`
- `Content-Type: application/json`

**Request Body:**
```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "admin": true
}
```

**Response:** Updated user object

---

#### `DELETE /api/v1/admin/users/:id`
Delete a user.

**Headers:**
- `Authorization: Bearer ADMIN_TOKEN`

**Response:**
```json
{
  "message": "User deleted"
}
```

---

### Coverage Upload (Coveralls Compatible)

#### `POST /upload/v2`
Upload coverage data in Coveralls format.

**Headers:**
- `Content-Type: application/json`

**Request Body:**
```json
{
  "repo_token": "PROJECT_TOKEN",
  "service_name": "github-actions",
  "service_number": "42",
  "service_job_id": "123456",
  "git": {
    "head": {
      "id": "abc123",
      "message": "Commit message"
    },
    "branch": "main"
  },
  "source_files": [
    {
      "name": "src/main.go",
      "source": "package main...",
      "coverage": [1, 2, 0, 1, null]
    }
  ]
}
```

**Response:**
```json
{
  "message": "Not implemented yet"
}
```

---

### Webhooks

#### `POST /webhook`
Receive webhooks from external services.

**Headers:**
- `Content-Type: application/json`

**Request Body:** Depends on webhook provider

**Response:**
```json
{
  "message": "Not implemented yet"
}
```

---

### Badges

#### `GET /projects/:id/badge.svg`
Get a coverage badge for a project.

**Response:** SVG image

Example:
```
![Coverage](http://localhost:4000/projects/1/badge.svg)
```

---

## Error Responses

All endpoints return standard error responses:

### 400 Bad Request
```json
{
  "error": "Invalid input"
}
```

### 401 Unauthorized
```json
{
  "error": "Authorization token required"
}
```

### 403 Forbidden
```json
{
  "error": "Admin privileges required"
}
```

### 404 Not Found
```json
{
  "error": "Resource not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error"
}
```

---

## Rate Limiting

Currently, there is no rate limiting implemented. This may be added in future versions.

## Pagination

Pagination is not yet implemented but will be added for list endpoints in future versions.

## Versioning

The API is versioned via the URL path (`/api/v1`). Major breaking changes will increment the version number.
