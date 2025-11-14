# Build stage for backend
FROM golang:1.25-alpine AS backend-builder

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source
COPY backend/ ./backend/

# Copy generated Swagger docs
COPY docs/ ./docs/

# Set Go build cache location
ENV GOCACHE=/root/.cache/go-build
# Build backend
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -o /librecov-server backend/cmd/server/main.go

# Build stage for frontend
FROM node:25-alpine AS frontend-builder

WORKDIR /app

# Copy frontend package files
COPY frontend/package*.json ./
RUN npm install

# Copy frontend source
COPY frontend/ ./

# Build frontend
RUN npm run build

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /librecov-server .

# Copy Swagger docs
COPY --from=backend-builder /app/docs ./docs

# Copy frontend build
COPY --from=frontend-builder /app/dist ./frontend/dist

# Create directory for static files
RUN mkdir -p /app/static

EXPOSE 4000

ENV PORT=4000

CMD ["./librecov-server"]
