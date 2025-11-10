# Build stage for backend
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source
COPY backend/ ./backend/

# Build backend
RUN CGO_ENABLED=0 GOOS=linux go build -o /librecov-server backend/cmd/server/main.go

# Build stage for frontend
FROM node:20-alpine AS frontend-builder

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

# Copy frontend build
COPY --from=frontend-builder /app/dist ./frontend/dist

# Create directory for static files
RUN mkdir -p /app/static

EXPOSE 4000

ENV PORT=4000

CMD ["./librecov-server"]
