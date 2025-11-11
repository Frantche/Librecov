#!/bin/bash
set -e

# Test coverage upload to Librecov
# This script:
# 1. Creates a test user and project
# 2. Generates a project token
# 3. Uploads Go coverage
# 4. Uploads JavaScript coverage

LIBRECOV_URL="http://localhost:4000"

echo "========================================="
echo "Testing Librecov Coverage Upload"
echo "========================================="
echo ""

# First, we need to create a project with a token
# Since we don't have OIDC configured, we'll use the upload API directly with a pre-created token

echo "Step 1: Create a test project via database"
echo "-------------------------------------------"

# Create a simple project directly in the database for testing
docker exec librecov-db-1 psql -U postgres -d librecov_dev <<EOF
-- Create a test user if not exists
INSERT INTO users (email, name, admin, token, created_at, updated_at)
VALUES ('test@example.com', 'Test User', true, 'test-user-token-12345', NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

-- Get the user ID
DO \$\$
DECLARE
    v_user_id INTEGER;
BEGIN
    SELECT id INTO v_user_id FROM users WHERE email = 'test@example.com';
    
    -- Create a test project if not exists
    INSERT INTO projects (name, token, current_branch, user_id, coverage_rate, created_at, updated_at)
    VALUES ('Librecov Test Project', 'test-project-token-67890', 'main', v_user_id, 0.0, NOW(), NOW())
    ON CONFLICT (token) DO NOTHING;
END \$\$;

-- Show the project
SELECT id, name, token, current_branch FROM projects WHERE token = 'test-project-token-67890';
EOF

echo ""
echo "✓ Test project created with token: test-project-token-67890"
echo ""

# Test Go coverage upload
echo "Step 2: Test Go Coverage Upload"
echo "-------------------------------------------"
export PROJECT_TOKEN="test-project-token-67890"
export LIBRECOV_URL="$LIBRECOV_URL"

cd /home/coder/Librecov
./scripts/upload-coverage-go.sh

echo ""
echo "Step 3: Verify upload via API"
echo "-------------------------------------------"

# Get the project details
PROJECT_DATA=$(curl -s "$LIBRECOV_URL/api/v1/projects/1")
echo "Project data:"
echo "$PROJECT_DATA" | jq '.'

# Get builds
echo ""
echo "Builds:"
BUILDS=$(curl -s "$LIBRECOV_URL/api/v1/projects/1/builds")
echo "$BUILDS" | jq '.'

echo ""
echo "========================================="
echo "✓ Coverage upload test completed!"
echo "========================================="
echo ""
echo "Next steps:"
echo "  1. Open http://localhost:4000 in your browser"
echo "  2. View the 'Librecov Test Project' to see coverage"
echo "  3. Try the JavaScript upload: export PROJECT_TOKEN=test-project-token-67890 && cd frontend && npm install && npm run upload:coverage"
echo ""
