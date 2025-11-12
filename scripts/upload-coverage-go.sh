#!/bin/bash
set -e

# Upload Go coverage to Librecov
# Usage: ./upload-coverage-go.sh

# Configuration
LIBRECOV_URL="${LIBRECOV_URL:-http://localhost:4000}"
PROJECT_TOKEN="${PROJECT_TOKEN:-}"
COVERAGE_FILE="${COVERAGE_FILE:-coverage.out}"

if [ -z "$PROJECT_TOKEN" ]; then
    echo "Error: PROJECT_TOKEN environment variable is required"
    echo "Usage: PROJECT_TOKEN=your-token ./upload-coverage-go.sh"
    exit 1
fi

# Change to backend directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")/backend"
cd "$BACKEND_DIR"

echo "==> Running Go tests with coverage..."
go test -v -covermode=count -coverprofile="$COVERAGE_FILE" ./... 2>&1 | grep -v "go: no such tool"

echo ""
echo "==> Coverage summary:"
go tool cover -func="$COVERAGE_FILE" | tail -5

echo ""
echo "==> Installing goveralls..."
go install github.com/mattn/goveralls@latest

echo ""
echo "==> Uploading coverage to Librecov at $LIBRECOV_URL..."

# Upload using goveralls with custom endpoint
goveralls -endpoint="$LIBRECOV_URL" -coverprofile="$COVERAGE_FILE" -service=manual -repotoken="$PROJECT_TOKEN"

echo ""
echo "Done!"
