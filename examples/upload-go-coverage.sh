#!/bin/bash
set -e

# Quick example: Upload Go coverage to Librecov
# 
# This is a complete working example using the Librecov project itself

echo "========================================="
echo "Librecov Coverage Upload Example"
echo "========================================="
echo ""

# Step 1: Set your project token
# Get this from your Librecov project settings page
if [ -z "$PROJECT_TOKEN" ]; then
    echo "Error: Please set PROJECT_TOKEN environment variable"
    echo "Example: export PROJECT_TOKEN=your-token-here"
    exit 1
fi

# Step 2: Set Librecov URL (optional, defaults to localhost)
LIBRECOV_URL="${LIBRECOV_URL:-http://localhost:4000}"

echo "Configuration:"
echo "  Librecov URL: $LIBRECOV_URL"
echo "  Project Token: ${PROJECT_TOKEN:0:10}..."
echo ""

# Step 3: Run tests with coverage
echo "Step 1: Running Go tests with coverage..."
cd "$(dirname "$0")/../backend"
go test -v -covermode=count -coverprofile=coverage.out ./... 2>&1 | grep -E "^(PASS|FAIL|ok|===)"

echo ""
echo "Step 2: Coverage summary..."
go tool cover -func=coverage.out | tail -5

echo ""
echo "Step 3: Uploading to Librecov..."
cd ..
./upload-coverage -coverprofile=backend/coverage.out

echo ""
echo "========================================="
echo "✓ Done! Check your project in Librecov:"
echo "  ${LIBRECOV_URL/\/upload\/v2/}"
echo "========================================="
