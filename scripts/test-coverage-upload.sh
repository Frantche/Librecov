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

# Wait for LibreCov application to be fully ready
echo "Waiting for LibreCov application to be ready..."
MAX_ATTEMPTS=30
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    if curl -s -f "$LIBRECOV_URL/health" > /dev/null 2>&1; then
        echo "✅ LibreCov application is ready!"
        break
    fi
    
    if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
        echo "❌ LibreCov application failed to start after $MAX_ATTEMPTS attempts"
        echo "Application logs:"
        docker logs librecov-librecov-1 2>/dev/null || echo "Could not retrieve logs"
        exit 1
    fi
    
    echo "Attempt $ATTEMPT/$MAX_ATTEMPTS - waiting for LibreCov..."
    sleep 3
    ATTEMPT=$((ATTEMPT + 1))
done

# Wait a bit more for migrations to complete
echo "Waiting for database migrations to complete..."
sleep 5

# First, verify database connection and tables exist
echo "Checking database connectivity..."
docker exec librecov-db-1 psql -U postgres -d librecov_dev -c "SELECT 1;" > /dev/null || {
    echo "❌ Cannot connect to database"
    exit 1
}
echo "✅ Database connection OK"

echo "Checking if required tables exist..."
TABLES_EXIST=$(docker exec librecov-db-1 psql -U postgres -d librecov_dev -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'projects');")
if [ "$TABLES_EXIST" -lt "2" ]; then
    echo "❌ Required tables (users, projects) do not exist. Waiting for migrations..."
    
    # Wait for migrations to complete
    MAX_MIGRATION_ATTEMPTS=20
    MIGRATION_ATTEMPT=1
    while [ $MIGRATION_ATTEMPT -le $MAX_MIGRATION_ATTEMPTS ]; do
        TABLES_EXIST=$(docker exec librecov-db-1 psql -U postgres -d librecov_dev -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'projects');")
        if [ "$TABLES_EXIST" -ge "2" ]; then
            echo "✅ Tables created after migration attempt $MIGRATION_ATTEMPT"
            break
        fi
        
        if [ $MIGRATION_ATTEMPT -eq $MAX_MIGRATION_ATTEMPTS ]; then
            echo "❌ Tables still don't exist after $MAX_MIGRATION_ATTEMPTS attempts"
            echo "Database tables:"
            docker exec librecov-db-1 psql -U postgres -d librecov_dev -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';"
            exit 1
        fi
        
        echo "Migration attempt $MIGRATION_ATTEMPT/$MAX_MIGRATION_ATTEMPTS - waiting..."
        sleep 5
        MIGRATION_ATTEMPT=$((MIGRATION_ATTEMPT + 1))
    done
else
    echo "✅ Required tables exist"
fi

# Create a simple project directly in the database for testing
echo "Creating test user and project..."

# Create user first
docker exec librecov-db-1 psql -U postgres -d librecov_dev -c "
INSERT INTO users (email, name, admin, token, created_at, updated_at)
VALUES ('test@example.com', 'Test User', true, 'test-user-token-12345', NOW(), NOW())
ON CONFLICT (email) DO NOTHING;
" || {
    echo "❌ Failed to create user"
    exit 1
}

# Get user ID
USER_ID=$(docker exec librecov-db-1 psql -U postgres -d librecov_dev -t -c "SELECT id FROM users WHERE email = 'test@example.com';")
USER_ID=$(echo "$USER_ID" | tr -d ' ')

if [ -z "$USER_ID" ]; then
    echo "❌ Could not get user ID"
    exit 1
fi

echo "User ID: $USER_ID"

# Create project if it doesn't exist
PROJECT_EXISTS=$(docker exec librecov-db-1 psql -U postgres -d librecov_dev -t -c "SELECT COUNT(*) FROM projects WHERE token = 'test-project-token-67890';")
PROJECT_EXISTS=$(echo "$PROJECT_EXISTS" | tr -d ' ')

if [ "$PROJECT_EXISTS" -eq "0" ]; then
    echo "Creating new test project..."
    PROJECT_ID=$(docker exec librecov-db-1 psql -U postgres -d librecov_dev -t -c "SELECT gen_random_uuid()::text;")
    PROJECT_ID=$(echo "$PROJECT_ID" | tr -d ' ')
    
    docker exec librecov-db-1 psql -U postgres -d librecov_dev -c "
    INSERT INTO projects (id, name, token, current_branch, user_id, coverage_rate, created_at, updated_at)
    VALUES ('$PROJECT_ID', 'Librecov Test Project', 'test-project-token-67890', 'main', $USER_ID, 0.0, NOW(), NOW());
    " || {
        echo "❌ Failed to create project"
        exit 1
    }
    echo "✅ Created project with ID: $PROJECT_ID"
else
    echo "✅ Project already exists"
fi

echo ""
echo "✓ Test project setup completed with token: test-project-token-67890"
echo ""

# Verify the project was created
echo "Verifying project creation..."
PROJECT_EXISTS=$(docker exec librecov-db-1 psql -U postgres -d librecov_dev -t -c "SELECT COUNT(*) FROM projects WHERE token = 'test-project-token-67890';")
if [ "$PROJECT_EXISTS" -eq "0" ]; then
    echo "❌ Project was not created successfully"
    echo "Checking what projects exist:"
    docker exec librecov-db-1 psql -U postgres -d librecov_dev -c "SELECT id, name, token FROM projects LIMIT 5;"
    exit 1
fi
echo "✅ Project verified in database"
echo ""

# Test Go coverage upload
echo "Step 2: Test Go Coverage Upload"
echo "-------------------------------------------"
export PROJECT_TOKEN="test-project-token-67890"
export LIBRECOV_URL="$LIBRECOV_URL"

"$(dirname "${BASH_SOURCE[0]}")/upload-coverage-go.sh"

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
echo "Step 4: Test authenticated API endpoints"
echo "-------------------------------------------"

# Get the latest build ID for testing
BUILD_ID=$(docker exec librecov-db-1 psql -U postgres -d librecov_dev -t -c "
SELECT b.id
FROM builds b
JOIN projects p ON p.id = b.project_id
WHERE p.token = 'test-project-token-67890'
ORDER BY b.created_at DESC
LIMIT 1;")

BUILD_ID=$(echo "$BUILD_ID" | tr -d ' ')

# Get the project ID for testing
PROJECT_ID=$(docker exec librecov-db-1 psql -U postgres -d librecov_dev -t -c "SELECT id FROM projects WHERE token = 'test-project-token-67890';")
PROJECT_ID=$(echo "$PROJECT_ID" | tr -d ' ')

if [ -n "$BUILD_ID" ]; then
    echo "Testing build API endpoint (should require authentication)..."
    BUILD_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$LIBRECOV_URL/api/v1/builds/$BUILD_ID")
    HTTP_STATUS=$(echo "$BUILD_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
    RESPONSE_BODY=$(echo "$BUILD_RESPONSE" | sed '/HTTP_STATUS:/d')
    
    if [ "$HTTP_STATUS" -eq "401" ]; then
        echo "✅ Build endpoint correctly requires authentication (401 Unauthorized)"
    else
        echo "❌ Build endpoint should require authentication but got status $HTTP_STATUS"
        echo "Response: $RESPONSE_BODY"
    fi
    
    echo "Testing project builds endpoint (should require authentication)..."
    PROJECT_BUILDS_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$LIBRECOV_URL/api/v1/projects/$PROJECT_ID/builds")
    HTTP_STATUS=$(echo "$PROJECT_BUILDS_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
    
    if [ "$HTTP_STATUS" -eq "401" ]; then
        echo "✅ Project builds endpoint correctly requires authentication (401 Unauthorized)"
    else
        echo "❌ Project builds endpoint should require authentication but got status $HTTP_STATUS"
    fi
else
    echo "⚠️ No build found to test authenticated endpoints"
fi

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
