# Coverage Upload - Quick Reference

This document provides quick commands for uploading coverage to Librecov.

## Prerequisites

1. **Librecov running** (Docker Compose or deployed)
2. **Project created** in Librecov UI
3. **Project token** from Settings page

## Go Projects

### Method 1: Using upload-coverage tool (Recommended)

```bash
# Build the tool (one time)
cd tools/upload-coverage
go build -o ../../upload-coverage .
cd ../..

# Run tests and upload
export PROJECT_TOKEN="your-token-here"
go test -covermode=count -coverprofile=coverage.out ./...
./upload-coverage -coverprofile=coverage.out
```

### Method 2: Using the example script

```bash
export PROJECT_TOKEN="your-token-here"
./examples/upload-go-coverage.sh
```

### Method 3: Using goveralls

```bash
go install github.com/mattn/goveralls@latest
export COVERALLS_ENDPOINT="http://localhost:4000/upload/v2"
export COVERALLS_REPO_TOKEN="your-token-here"

go test -covermode=count -coverprofile=coverage.out ./...
goveralls -service=manual -repotoken=$COVERALLS_REPO_TOKEN -coverprofile=coverage.out
```

## JavaScript/Node.js Projects

### Setup (one time)

```bash
cd frontend

# Install test dependencies
npm install -D vitest @vitest/coverage-v8

# Add to package.json scripts:
# "test": "vitest",
# "test:coverage": "vitest run --coverage"
```

### Run and upload

```bash
export PROJECT_TOKEN="your-token-here"
npm run test:coverage
node ../scripts/upload-coverage-js.js
```

## Testing This Repository

This repository includes complete working examples:

### Test Go Backend Coverage

```bash
# Get your project token from Librecov UI
# Then run:

cd /path/to/Librecov
export PROJECT_TOKEN="your-actual-token"
./examples/upload-go-coverage.sh
```

Expected output:
```
=========================================
Librecov Coverage Upload Example
=========================================

Configuration:
  Librecov URL: http://localhost:4000
  Project Token: your-token...

Step 1: Running Go tests with coverage...
PASS
ok      github.com/Frantche/Librecov/backend/internal/api       0.028s  coverage: 10.9% of statements
...

Step 2: Coverage summary...
total:  (statements)  14.1%

Step 3: Uploading to Librecov...
==> Parsing coverage profile...
   Processed 8 files
   Overall coverage: 15.03% (165/1098 lines)
   Branch: feat/example-go-javascript, Commit: ff05c7e6

==> Uploading to http://localhost:4000/upload/v2...
{
  "build_id": 2,
  "coverage_rate": 15.027322404371585,
  "job_id": 2,
  "message": "Coverage uploaded successfully",
  "project_id": 1
}

✓ Coverage uploaded successfully!
```

### Verify Upload

1. Open http://localhost:4000 in your browser
2. View your project
3. Check the builds list - you should see a new build with ~15% coverage
4. Click on the build to see file-level coverage

## API Endpoint

```
POST /upload/v2
Content-Type: application/json

{
  "repo_token": "your-project-token",
  "service_name": "manual",
  "git": {
    "branch": "main",
    "head": {
      "id": "commit-sha",
      "message": "commit message"
    }
  },
  "source_files": [
    {
      "name": "path/to/file.go",
      "source": "file contents...",
      "coverage": [null, 1, 1, 0, null, 1]
    }
  ]
}
```

Coverage array format:
- `null` = not executable (comment, blank line)
- `0` = executable but not covered
- `> 0` = covered (number = hit count)

## Troubleshooting

### "Invalid repo token"
- Check token is correct
- Ensure project exists in Librecov
- Use project token, not user token

### "upload failed (status 401)"
- Token is invalid or expired
- Generate new token from project settings

### No coverage data uploaded
- Check coverage file exists and is not empty
- Verify coverage file format: `head coverage.out`
- Should start with: `mode: count`

### Wrong coverage numbers
- Ensure tests ran successfully
- Check that all source files are in the coverage report
- Verify module path in go.mod matches file paths

## CI/CD Integration

### GitHub Actions

```yaml
- name: Upload coverage
  env:
    PROJECT_TOKEN: ${{ secrets.LIBRECOV_PROJECT_TOKEN }}
  run: |
    go test -covermode=count -coverprofile=coverage.out ./...
    ./upload-coverage -coverprofile=coverage.out
```

### GitLab CI

```yaml
test:
  script:
    - go test -covermode=count -coverprofile=coverage.out ./...
    - ./upload-coverage -coverprofile=coverage.out
  variables:
    PROJECT_TOKEN: $LIBRECOV_PROJECT_TOKEN
```

### Drone CI

```yaml
- name: coverage
  environment:
    PROJECT_TOKEN:
      from_secret: librecov_project_token
  commands:
    - go test -covermode=count -coverprofile=coverage.out ./...
    - ./upload-coverage -coverprofile=coverage.out
```

## Full Documentation

See [COVERAGE_UPLOAD.md](./COVERAGE_UPLOAD.md) for complete documentation.
