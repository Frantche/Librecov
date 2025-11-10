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
- (Optional) Kubernetes 1.19+ and Helm 3.0+ for Kubernetes deployment

### Installation

#### Option 1: Kubernetes with Helm (Recommended for Production)

1. **Install from OCI registry**
   ```bash
   helm install librecov oci://ghcr.io/frantche/charts/librecov --version 1.0.0
   ```

2. **Or install from source**
   ```bash
   git clone https://github.com/Frantche/Librecov.git
   cd Librecov
   helm install librecov ./helm/librecov
   ```

3. **With custom values**
   ```bash
   helm install librecov oci://ghcr.io/frantche/charts/librecov \
     --set ingress.enabled=true \
     --set ingress.hosts[0].host=librecov.example.com
   ```

See [Helm Chart README](./helm/librecov/README.md) for detailed configuration options.

#### Option 2: Docker Compose (Quick Local Development)

1. **Clone the repository**
   ```bash
   git clone https://github.com/Frantche/Librecov.git
   cd Librecov
   ```

2. **Start with Docker Compose**
   ```bash
   docker-compose up
   ```
   
   Access the application at http://localhost:4000

#### Option 3: Manual Installation

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

LibreCov supports flexible authentication:

1. **OIDC (OpenID Connect)**: Configure OIDC environment variables to enable SSO for user login
2. **API Tokens**: Use project tokens for uploading coverage data

#### OIDC Authentication Flow

When OIDC is configured, the frontend automatically detects it and provides SSO login:

1. The frontend fetches authentication configuration from `/auth/config`
2. If OIDC is enabled, users are redirected to the OIDC provider for login
3. After successful authentication, users are redirected back with an access token
4. The token is used for all subsequent API requests

If OIDC is not configured, the login page will display instructions for setting up OIDC authentication.

**OIDC Configuration:**
```env
OIDC_ISSUER=https://your-oidc-provider.com
OIDC_CLIENT_ID=your-client-id
OIDC_CLIENT_SECRET=your-client-secret
OIDC_REDIRECT_URL=http://localhost:4000/auth/callback
```

**Note:** The backend serves the frontend in production mode. When you build the Docker image or run `make build`, the frontend is built and served by the backend from the `/` route. All frontend routes are handled by the SPA, and API routes are available under `/api/v1`.

## Uploading Coverage

LibreCov accepts coverage data in Coveralls format. Most coverage tools that support Coveralls will work with LibreCov.

Example with `goveralls`:
```bash
goveralls -service=librecov -repotoken=YOUR_PROJECT_TOKEN -coverprofile=coverage.out
```

Set the endpoint URL to your LibreCov instance.

## Deployment

### Kubernetes with Helm

LibreCov can be deployed to Kubernetes using the official Helm chart:

```bash
# Install from OCI registry
helm install librecov oci://ghcr.io/frantche/charts/librecov --version 1.0.0

# Or from source
helm install librecov ./helm/librecov
```

**Custom configuration:**
```bash
# With ingress enabled
helm install librecov oci://ghcr.io/frantche/charts/librecov \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=librecov.example.com \
  --set ingress.className=nginx

# With OIDC authentication
helm install librecov oci://ghcr.io/frantche/charts/librecov \
  --set config.oidc.enabled=true \
  --set config.oidc.issuer=https://your-oidc-provider.com \
  --set config.oidc.clientId=your-client-id \
  --set config.oidc.clientSecret=your-client-secret

# With external PostgreSQL
helm install librecov oci://ghcr.io/frantche/charts/librecov \
  --set postgresql.enabled=false \
  --set externalDatabase.host=postgres.example.com \
  --set externalDatabase.password=yourpassword
```

For more configuration options, see the [Helm Chart README](./helm/librecov/README.md).

### Docker

Build and run with Docker:

```bash
# Build the image
docker build -t librecov:latest .

# Run with Docker Compose
docker-compose up
```

The Docker Compose setup includes PostgreSQL and is ready for development or testing.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

This project is a reimplementation of the original [LibreCov](https://github.com/Librecov/librecov) built with Elixir/Phoenix, now rebuilt with Golang and Vue.js.