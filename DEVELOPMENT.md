# Development Guide

## Getting Started

### Prerequisites

- Go 1.24 or higher
- Node.js 20 or higher
- PostgreSQL 12 or higher
- Make (optional, for convenience commands)

### Setting Up the Development Environment

1. **Clone the repository**
   ```bash
   git clone https://github.com/Frantche/Librecov.git
   cd Librecov
   ```

2. **Install dependencies**
   ```bash
   make install
   ```
   
   Or manually:
   ```bash
   go mod download
   cd frontend && npm install
   ```

3. **Set up the database**
   ```bash
   createdb librecov_dev
   ```

4. **Configure environment variables**
   Copy the example environment file and update as needed:
   ```bash
   cp .env.example .env
   ```

5. **Start the development servers**
   
   In one terminal, start the backend:
   ```bash
   make dev-backend
   ```
   
   In another terminal, start the frontend:
   ```bash
   make dev-frontend
   ```
   
   Or manually:
   ```bash
   # Terminal 1 - Backend
   go run backend/cmd/server/main.go
   
   # Terminal 2 - Frontend
   cd frontend && npm run dev
   ```

### Project Structure

```
Librecov/
├── backend/                    # Golang backend
│   ├── cmd/
│   │   └── server/            # Main application entry point
│   │       └── main.go
│   └── internal/              # Private application code
│       ├── api/               # API handlers and routes
│       ├── auth/              # Authentication (OIDC)
│       ├── database/          # Database connection and migrations
│       ├── middleware/        # HTTP middleware
│       ├── models/            # Data models
│       └── services/          # Business logic
├── frontend/                   # Vue.js frontend
│   ├── src/
│   │   ├── components/        # Reusable Vue components
│   │   ├── views/             # Page components
│   │   ├── stores/            # Pinia stores
│   │   ├── services/          # API client
│   │   ├── router/            # Vue Router configuration
│   │   └── types/             # TypeScript types
│   └── public/                # Static assets
├── .github/                    # GitHub Actions workflows
├── Dockerfile                  # Docker build configuration
├── docker-compose.yml          # Docker Compose setup
├── Makefile                    # Build commands
└── README.md                   # Project documentation
```

## Development Workflow

### Running Tests

**Backend:**
```bash
make test-backend
```

Or manually:
```bash
go test -v ./backend/...
```

With coverage:
```bash
make test-coverage
```

**Frontend:**
```bash
cd frontend && npm test
```

### Building

**Backend:**
```bash
make build
```

Or manually:
```bash
go build -o bin/librecov-server backend/cmd/server/main.go
```

**Frontend:**
```bash
cd frontend && npm run build
```

### Linting

**Backend:**
```bash
make lint
```

**Frontend:**
```bash
cd frontend && npm run lint
```

## OIDC Configuration

LibreCov supports OpenID Connect (OIDC) for authentication. To enable OIDC:

1. Set up an OIDC provider (e.g., Keycloak, Auth0, Okta)
2. Configure the environment variables in `.env`:
   ```env
   OIDC_ISSUER=https://your-oidc-provider.com
   OIDC_CLIENT_ID=your-client-id
   OIDC_REDIRECT_URL=http://localhost:4000/auth/callback
   ```

3. Restart the backend server

### OIDC Flow

1. User clicks "Login" button
2. User is redirected to OIDC provider
3. User authenticates with OIDC provider
4. OIDC provider redirects back to `/auth/callback` with authorization code
5. Backend exchanges code for ID token
6. Backend creates or updates user based on OIDC claims
7. Backend returns user info and token to frontend
8. Frontend stores token and uses it for API requests

## Database Migrations

Currently, migrations are handled by GORM's AutoMigrate feature. When you start the application, it will automatically create or update the database schema.

Future: We plan to add a proper migration system for production use.

## API Documentation

The API is RESTful and mostly compatible with the Coveralls API.

### Authentication

Most endpoints require authentication via Bearer token:
```
Authorization: Bearer <token>
```

### Key Endpoints

- `GET /api/v1/projects` - List projects
- `POST /api/v1/projects` - Create project
- `GET /api/v1/projects/:id` - Get project details
- `GET /api/v1/projects/:projectId/builds` - List builds for project
- `GET /api/v1/builds/:id` - Get build details
- `POST /upload/v2` - Upload coverage data (Coveralls format)
- `GET /projects/:id/badge.svg` - Get coverage badge

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Run tests and linting
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## Troubleshooting

### Database Connection Issues

If you see database connection errors:
- Ensure PostgreSQL is running
- Check the database credentials in `.env`
- Verify the database exists: `psql -l | grep librecov`

### Frontend Build Issues

If frontend build fails:
- Clear node_modules: `rm -rf frontend/node_modules`
- Reinstall: `cd frontend && npm install`
- Clear cache: `cd frontend && npm cache clean --force`

### OIDC Issues

If OIDC login fails:
- Verify OIDC configuration in `.env`
- Check OIDC provider logs
- Ensure redirect URL is configured correctly in OIDC provider
- Check browser console for errors
