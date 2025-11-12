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

-- Get the user ID and create a test project
DO \$\$
DECLARE
    v_user_id INTEGER;
    v_project_id VARCHAR(36);
BEGIN
    SELECT id INTO v_user_id FROM users WHERE email = 'test@example.com';
    
    -- Generate a UUID for the project
    v_project_id := gen_random_uuid()::text;
    
    -- Create a test project if not exists
    INSERT INTO projects (id, name, token, current_branch, user_id, coverage_rate, created_at, updated_at)
    VALUES (v_project_id, 'Librecov Test Project', 'test-project-token-67890', 'main', v_user_id, 0.0, NOW(), NOW())
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

# Get the project details by querying the database directly
echo "Step 3: Verify upload via database"
echo "-------------------------------------------"

PROJECT_INFO=$(docker exec librecov-db-1 psql -U postgres -d librecov_dev -t -c "
SELECT 
    p.id, p.name, p.token, p.coverage_rate,
    COUNT(b.id) as build_count,
    COUNT(j.id) as job_count
FROM projects p
LEFT JOIN builds b ON b.project_id = p.id
LEFT JOIN jobs j ON j.build_id = b.id
WHERE p.token = 'test-project-token-67890'
GROUP BY p.id, p.name, p.token, p.coverage_rate;")

echo "Project info from database:"
echo "$PROJECT_INFO"

# Also check builds and jobs
echo ""
echo "Recent builds:"
docker exec librecov-db-1 psql -U postgres -d librecov_dev -c "
SELECT b.id, b.build_num, b.branch, b.commit_sha, b.coverage_rate, b.created_at
FROM builds b
JOIN projects p ON p.id = b.project_id
WHERE p.token = 'test-project-token-67890'
ORDER BY b.created_at DESC
LIMIT 5;"

echo ""
echo "Recent jobs:"
docker exec librecov-db-1 psql -U postgres -d librecov_dev -c "
SELECT j.id, j.job_number, j.coverage_rate, b.build_num, p.name
FROM jobs j
JOIN builds b ON b.id = j.build_id
JOIN projects p ON p.id = b.project_id
WHERE p.token = 'test-project-token-67890'
ORDER BY j.created_at DESC
LIMIT 5;"

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
