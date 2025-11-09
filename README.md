# LibreCov

LibreCov is a self-hosted open-source code coverage history viewer built with Golang and Vue.js. It provides a web interface to track and visualize test coverage over time, and is mostly compatible with the Coveralls API.

## Features

- 📊 **Coverage History Tracking**: View coverage trends across builds
- 🔐 **OIDC Authentication**: Secure authentication using OpenID Connect
- 📁 **Project Management**: Organize multiple projects with coverage data
- 🏗️ **Build Tracking**: Track coverage for each build and commit
- 📄 **File-level Coverage**: View detailed coverage for individual files
- 👥 **User Management**: Admin panel for managing users and permissions
- 🎯 **Coveralls Compatible**: Works with existing Coveralls-compatible tools

## Tech Stack

### Backend
- **Golang** with Gin web framework
- **PostgreSQL** database with GORM
- **OIDC** authentication support
- RESTful API design

### Frontend
- **Vue.js 3** with Composition API
- **TypeScript** for type safety
- **Vite** for fast development and building
- **Pinia** for state management
- **Vue Router** for navigation

## Quick Start

### Prerequisites

- Go 1.24+ 
- Node.js 20+
- PostgreSQL 12+
- (Optional) OIDC provider for authentication

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/Frantche/Librecov.git
   cd Librecov
   ```

2. **Set up the database**
   ```bash
   createdb librecov_dev
   ```

3. **Configure environment variables**
   Create a `.env` file in the root directory:
   ```env
   # Database
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=yourpassword
   DB_NAME=librecov_dev
   DB_SSLMODE=disable

   # Server
   PORT=4000

   # OIDC (optional)
   OIDC_ISSUER=https://your-oidc-provider.com
   OIDC_CLIENT_ID=your-client-id
   OIDC_CLIENT_SECRET=your-client-secret
   OIDC_REDIRECT_URL=http://localhost:4000/auth/callback
   ```

4. **Install backend dependencies**
   ```bash
   go mod download
   ```

5. **Install frontend dependencies**
   ```bash
   cd frontend
   npm install
   cd ..
   ```

### Development

Run backend and frontend concurrently:

**Terminal 1 - Backend:**
```bash
go run backend/cmd/server/main.go
```

**Terminal 2 - Frontend:**
```bash
cd frontend
npm run dev
```

The application will be available at:
- Frontend: http://localhost:3000
- Backend API: http://localhost:4000

### Building for Production

**Build backend:**
```bash
go build -o librecov-server backend/cmd/server/main.go
```

**Build frontend:**
```bash
cd frontend
npm run build
cd ..
```

The frontend build will be in `frontend/dist/`.

### Running Tests

**Backend tests:**
```bash
go test ./...
```

**Frontend tests:**
```bash
cd frontend
npm run test
```

## API Documentation

The API is mostly compatible with the Coveralls API. Key endpoints:

- `POST /upload/v2` - Upload coverage data (Coveralls compatible)
- `GET /api/v1/projects` - List projects
- `GET /api/v1/projects/:id` - Get project details
- `GET /api/v1/builds/:id` - Get build details
- `GET /api/v1/jobs/:id` - Get job details
- `GET /projects/:id/badge.svg` - Get coverage badge

### Authentication

LibreCov supports two authentication methods:

1. **OIDC (OpenID Connect)**: Configure OIDC environment variables to enable SSO
2. **API Tokens**: Use project tokens for uploading coverage data

## Uploading Coverage

LibreCov accepts coverage data in Coveralls format. Most coverage tools that support Coveralls will work with LibreCov.

Example with `goveralls`:
```bash
goveralls -service=librecov -repotoken=YOUR_PROJECT_TOKEN -coverprofile=coverage.out
```

Set the endpoint URL to your LibreCov instance.

## Docker Deployment

Coming soon!

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

This project is a reimplementation of the original [LibreCov](https://github.com/Librecov/librecov) built with Elixir/Phoenix, now rebuilt with Golang and Vue.js.