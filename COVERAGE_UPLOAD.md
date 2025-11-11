# Coverage Upload Guide

This guide shows how to upload code coverage reports to Librecov from **Go** and **JavaScript/Node.js** projects.

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Getting Your Project Token](#getting-your-project-token)
3. [Go Coverage Upload](#go-coverage-upload)
4. [JavaScript/Node.js Coverage Upload](#javascriptnodejs-coverage-upload)
5. [CI/CD Integration](#cicd-integration)
6. [Troubleshooting](#troubleshooting)

---

## Prerequisites

1. **Running Librecov instance** (via Docker Compose or deployed)
2. **Authentication**: Login via OIDC or have a user API token
3. **Project created** in Librecov UI
4. **Project token** (from project settings page)

---

## Getting Your Project Token

### Option 1: Via Web UI
1. Login to Librecov
2. Navigate to your project
3. Click "Settings" button
4. Copy a project token from the "Project Tokens" section
5. Or create a new token with a descriptive name

### Option 2: Via API
```bash
# Login and get session cookie first, then:
curl -X GET http://localhost:4000/api/v1/projects/1/tokens \
  -H "Cookie: session_id=YOUR_SESSION_ID"

# Create a new project token:
curl -X POST http://localhost:4000/api/v1/projects/1/tokens \
  -H "Cookie: session_id=YOUR_SESSION_ID" \
  -H "Content-Type: application/json" \
  -d '{"name": "CI Token"}'
```

### Option 3: Legacy Project Token
Each project has a default token shown in the project list. This is the `token` field in the project object.

---

## Go Coverage Upload

### Step 1: Generate Coverage Report

Librecov accepts the **Coveralls JSON format**. For Go projects, we need to convert Go coverage to this format.

#### Install goveralls (recommended)
```bash
go install github.com/mattn/goveralls@latest
```

#### Generate Coverage
```bash
# Run tests with coverage
go test -v -covermode=count -coverprofile=coverage.out ./...

# Convert to Coveralls JSON format
goveralls -coverprofile=coverage.out -service=local -repotoken=YOUR_PROJECT_TOKEN
```

### Step 2: Manual Upload Script (Alternative)

If you prefer manual control, create a script to upload coverage:

**upload-coverage-go.sh**:
```bash
#!/bin/bash
set -e

# Configuration
LIBRECOV_URL="${LIBRECOV_URL:-http://localhost:4000}"
PROJECT_TOKEN="${PROJECT_TOKEN:-}"
COVERAGE_FILE="${COVERAGE_FILE:-coverage.out}"

if [ -z "$PROJECT_TOKEN" ]; then
    echo "Error: PROJECT_TOKEN environment variable is required"
    exit 1
fi

# Generate coverage report
echo "Running tests and generating coverage..."
go test -v -covermode=count -coverprofile="$COVERAGE_FILE" ./...

# Install goveralls if not present
if ! command -v goveralls &> /dev/null; then
    echo "Installing goveralls..."
    go install github.com/mattn/goveralls@latest
fi

# Get git information
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_MESSAGE=$(git log -1 --pretty=%B 2>/dev/null || echo "No commit message")

# Convert coverage to Coveralls JSON
echo "Converting coverage to Coveralls format..."
COVERALLS_JSON=$(mktemp)
goveralls -coverprofile="$COVERAGE_FILE" -service=manual -repotoken="$PROJECT_TOKEN" -dryrun=true > "$COVERALLS_JSON" 2>/dev/null || true

# If goveralls doesn't support -dryrun, use alternative method
if [ ! -s "$COVERALLS_JSON" ]; then
    echo "Using go-cover-view for conversion..."
    # Alternative: manual conversion using go tool cover
    go tool cover -func="$COVERAGE_FILE" > /dev/null
    
    # Create JSON manually (simplified example)
    cat > "$COVERALLS_JSON" << EOF
{
  "repo_token": "$PROJECT_TOKEN",
  "service_name": "manual",
  "git": {
    "branch": "$GIT_BRANCH",
    "head": {
      "id": "$GIT_COMMIT",
      "message": "$GIT_MESSAGE"
    }
  },
  "source_files": []
}
EOF
fi

# Upload to Librecov
echo "Uploading coverage to Librecov..."
RESPONSE=$(curl -s -X POST "$LIBRECOV_URL/api/upload" \
    -H "Content-Type: application/json" \
    -d @"$COVERALLS_JSON")

echo "Response: $RESPONSE"

# Cleanup
rm -f "$COVERALLS_JSON"

echo "Coverage upload complete!"
```

Make it executable:
```bash
chmod +x upload-coverage-go.sh
```

### Step 3: Upload

```bash
# Set your project token
export PROJECT_TOKEN="your-project-token-here"
export LIBRECOV_URL="http://localhost:4000"

# Run the upload script
./upload-coverage-go.sh
```

### Step 4: Using a Custom Go Tool

For better integration, here's a minimal Go program to upload coverage:

**tools/upload-coverage/main.go**:
```go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type CoverallsUpload struct {
	RepoToken   string      `json:"repo_token"`
	ServiceName string      `json:"service_name"`
	Git         *GitInfo    `json:"git,omitempty"`
	SourceFiles []SourceFile `json:"source_files"`
}

type GitInfo struct {
	Branch string   `json:"branch"`
	Head   HeadInfo `json:"head"`
}

type HeadInfo struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type SourceFile struct {
	Name     string        `json:"name"`
	Source   string        `json:"source"`
	Coverage []interface{} `json:"coverage"`
}

func main() {
	token := flag.String("token", os.Getenv("PROJECT_TOKEN"), "Project token")
	url := flag.String("url", getEnvOrDefault("LIBRECOV_URL", "http://localhost:4000"), "Librecov URL")
	coverProfile := flag.String("coverprofile", "coverage.out", "Coverage profile file")
	flag.Parse()

	if *token == "" {
		fmt.Fprintf(os.Stderr, "Error: project token required (use -token or PROJECT_TOKEN env)\n")
		os.Exit(1)
	}

	// Read coverage file
	coverage, err := parseCoverageProfile(*coverProfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing coverage: %v\n", err)
		os.Exit(1)
	}

	// Get git info
	gitInfo := getGitInfo()

	// Build upload payload
	upload := CoverallsUpload{
		RepoToken:   *token,
		ServiceName: "manual",
		Git:         gitInfo,
		SourceFiles: coverage,
	}

	// Upload
	if err := uploadCoverage(*url, upload); err != nil {
		fmt.Fprintf(os.Stderr, "Error uploading coverage: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Coverage uploaded successfully!")
}

func parseCoverageProfile(filename string) ([]SourceFile, error) {
	// This is a simplified parser - use goveralls or a proper library in production
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	files := make(map[string]*SourceFile)
	lines := strings.Split(string(data), "\n")
	
	for _, line := range lines[1:] { // Skip first line (mode declaration)
		if line == "" {
			continue
		}
		
		// Parse: filename:startLine.startCol,endLine.endCol numStmt count
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		
		fileAndRange := strings.Split(parts[0], ":")
		if len(fileAndRange) < 2 {
			continue
		}
		
		filename := fileAndRange[0]
		if _, ok := files[filename]; !ok {
			source, _ := os.ReadFile(filename)
			files[filename] = &SourceFile{
				Name:     filename,
				Source:   string(source),
				Coverage: []interface{}{},
			}
		}
	}

	result := make([]SourceFile, 0, len(files))
	for _, file := range files {
		result = append(result, *file)
	}
	
	return result, nil
}

func getGitInfo() *GitInfo {
	branch := execGit("rev-parse", "--abbrev-ref", "HEAD")
	commit := execGit("rev-parse", "HEAD")
	message := execGit("log", "-1", "--pretty=%B")
	
	if branch == "" || commit == "" {
		return nil
	}
	
	return &GitInfo{
		Branch: branch,
		Head: HeadInfo{
			ID:      commit,
			Message: message,
		},
	}
}

func execGit(args ...string) string {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func uploadCoverage(baseURL string, upload CoverallsUpload) error {
	data, err := json.Marshal(upload)
	if err != nil {
		return err
	}

	resp, err := http.Post(baseURL+"/api/upload", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed (status %d): %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Response: %s\n", string(body))
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
```

---

## JavaScript/Node.js Coverage Upload

### Step 1: Setup Test Coverage

For a **Vite + Vitest** project (like the Librecov frontend):

**vitest.config.ts** (or **vite.config.ts**):
```typescript
import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    coverage: {
      provider: 'v8', // or 'istanbul'
      reporter: ['text', 'json', 'html', 'lcov'],
      reportsDirectory: './coverage',
      include: ['src/**/*.{js,ts,vue}'],
      exclude: ['node_modules', 'dist', '**/*.spec.ts', '**/*.test.ts']
    }
  }
})
```

Install dependencies:
```bash
npm install -D vitest @vitest/coverage-v8
# or
npm install -D vitest @vitest/coverage-istanbul
```

### Step 2: Generate Coverage

```bash
# Run tests with coverage
npm run test:coverage
# or
npx vitest run --coverage
```

This generates `coverage/coverage-final.json`.

### Step 3: Upload Script

**upload-coverage-js.sh**:
```bash
#!/bin/bash
set -e

# Configuration
LIBRECOV_URL="${LIBRECOV_URL:-http://localhost:4000}"
PROJECT_TOKEN="${PROJECT_TOKEN:-}"
COVERAGE_DIR="${COVERAGE_DIR:-coverage}"

if [ -z "$PROJECT_TOKEN" ]; then
    echo "Error: PROJECT_TOKEN environment variable is required"
    exit 1
fi

# Run tests with coverage
echo "Running tests and generating coverage..."
npm run test:coverage || npx vitest run --coverage

# Check if coverage file exists
if [ ! -f "$COVERAGE_DIR/coverage-final.json" ]; then
    echo "Error: Coverage file not found at $COVERAGE_DIR/coverage-final.json"
    exit 1
fi

# Get git information
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_MESSAGE=$(git log -1 --pretty=%B 2>/dev/null || echo "No commit message")

# Convert coverage to Coveralls format using Node.js
node - <<EOF
const fs = require('fs');
const path = require('path');

// Read coverage data
const coverageData = JSON.parse(fs.readFileSync('$COVERAGE_DIR/coverage-final.json', 'utf8'));

// Convert to Coveralls format
const sourceFiles = [];

for (const [filePath, fileData] of Object.entries(coverageData)) {
  // Read source file
  let source = '';
  try {
    source = fs.readFileSync(filePath, 'utf8');
  } catch (err) {
    console.warn('Could not read source file:', filePath);
    continue;
  }

  // Build coverage array (null for non-executable, 0 for not covered, >0 for covered)
  const lines = source.split('\\n');
  const coverage = new Array(lines.length).fill(null);

  // Map statement coverage to line coverage
  if (fileData.statementMap && fileData.s) {
    for (const [key, stmt] of Object.entries(fileData.statementMap)) {
      const line = stmt.start.line - 1;
      if (line >= 0 && line < coverage.length) {
        coverage[line] = fileData.s[key] || 0;
      }
    }
  }

  sourceFiles.push({
    name: filePath.replace(process.cwd() + '/', ''),
    source: source,
    coverage: coverage
  });
}

// Build Coveralls JSON
const coverallsJson = {
  repo_token: '$PROJECT_TOKEN',
  service_name: 'manual',
  git: {
    branch: '$GIT_BRANCH',
    head: {
      id: '$GIT_COMMIT',
      message: '$GIT_MESSAGE'
    }
  },
  source_files: sourceFiles
};

// Write to temp file
fs.writeFileSync('/tmp/coveralls.json', JSON.stringify(coverallsJson, null, 2));
console.log('Converted coverage for', sourceFiles.length, 'files');
EOF

# Upload to Librecov
echo "Uploading coverage to Librecov..."
RESPONSE=$(curl -s -X POST "$LIBRECOV_URL/api/upload" \
    -H "Content-Type: application/json" \
    -d @/tmp/coveralls.json)

echo "Response: $RESPONSE"

# Cleanup
rm -f /tmp/coveralls.json

echo "Coverage upload complete!"
```

Make it executable:
```bash
chmod +x upload-coverage-js.sh
```

### Step 4: Upload

```bash
export PROJECT_TOKEN="your-project-token-here"
export LIBRECOV_URL="http://localhost:4000"
./upload-coverage-js.sh
```

### Alternative: Use coveralls-node

```bash
npm install -D coveralls

# After running tests with coverage:
cat coverage/lcov.info | npx coveralls
```

Note: Configure `coveralls` to point to your Librecov instance by setting:
```bash
export COVERALLS_ENDPOINT="http://localhost:4000/api/upload"
export COVERALLS_REPO_TOKEN="your-project-token"
```

---

## CI/CD Integration

### GitHub Actions

**.github/workflows/coverage.yml**:
```yaml
name: Coverage

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  go-coverage:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests with coverage
        run: go test -v -covermode=count -coverprofile=coverage.out ./...
      
      - name: Upload coverage to Librecov
        env:
          PROJECT_TOKEN: ${{ secrets.LIBRECOV_PROJECT_TOKEN }}
          LIBRECOV_URL: ${{ secrets.LIBRECOV_URL }}
        run: |
          go install github.com/mattn/goveralls@latest
          goveralls -coverprofile=coverage.out -service=github -repotoken=$PROJECT_TOKEN

  js-coverage:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Run tests with coverage
        run: npm run test:coverage
      
      - name: Upload coverage to Librecov
        env:
          PROJECT_TOKEN: ${{ secrets.LIBRECOV_PROJECT_TOKEN }}
          LIBRECOV_URL: ${{ secrets.LIBRECOV_URL }}
        run: ./upload-coverage-js.sh
```

### GitLab CI

**.gitlab-ci.yml**:
```yaml
go-coverage:
  image: golang:1.21
  script:
    - go test -v -covermode=count -coverprofile=coverage.out ./...
    - go install github.com/mattn/goveralls@latest
    - goveralls -coverprofile=coverage.out -service=gitlab -repotoken=$PROJECT_TOKEN
  variables:
    PROJECT_TOKEN: $LIBRECOV_PROJECT_TOKEN
    LIBRECOV_URL: $LIBRECOV_URL

js-coverage:
  image: node:18
  script:
    - npm ci
    - npm run test:coverage
    - ./upload-coverage-js.sh
  variables:
    PROJECT_TOKEN: $LIBRECOV_PROJECT_TOKEN
    LIBRECOV_URL: $LIBRECOV_URL
```

### Drone CI

**.drone.yml**:
```yaml
kind: pipeline
name: coverage

steps:
  - name: go-coverage
    image: golang:1.21
    environment:
      PROJECT_TOKEN:
        from_secret: librecov_project_token
      LIBRECOV_URL:
        from_secret: librecov_url
    commands:
      - go test -v -covermode=count -coverprofile=coverage.out ./...
      - go install github.com/mattn/goveralls@latest
      - goveralls -coverprofile=coverage.out -service=drone -repotoken=$PROJECT_TOKEN

  - name: js-coverage
    image: node:18
    environment:
      PROJECT_TOKEN:
        from_secret: librecov_project_token
      LIBRECOV_URL:
        from_secret: librecov_url
    commands:
      - npm ci
      - npm run test:coverage
      - ./upload-coverage-js.sh
```

---

## Troubleshooting

### Error: "Invalid repo token"
- Verify your `PROJECT_TOKEN` is correct
- Check that the project exists in Librecov
- Ensure you're using a project token (not a user token)

### Error: "Invalid JSON format"
- Ensure your JSON follows the Coveralls format
- Check that `repo_token` field is present
- Validate JSON syntax: `cat coveralls.json | jq .`

### Coverage not appearing
- Check Librecov logs: `docker logs librecov-librecov-1`
- Verify the upload response indicates success
- Check project page in UI for new builds

### Network errors
- Ensure Librecov is accessible from your CI environment
- Check firewall/network policies
- Verify `LIBRECOV_URL` is correct

### Git information missing
- Ensure git is installed in CI environment
- Check that `.git` directory exists
- Set git info manually in CI if needed:
  ```bash
  export GIT_BRANCH=$CI_COMMIT_BRANCH
  export GIT_COMMIT=$CI_COMMIT_SHA
  ```

---

## API Reference

### Upload Endpoint

**POST** `/upload/v2`

**Headers**:
```
Content-Type: application/json
```

**Body** (Coveralls format):
```json
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
      "source": "file content...",
      "coverage": [null, 1, 1, 0, null, 1]
    }
  ]
}
```

**Response**:
```json
{
  "message": "Coverage uploaded successfully",
  "project_id": 1,
  "build_id": 5,
  "job_id": 10,
  "coverage_rate": 85.5
}
```

**Coverage Array Format**:
- `null`: Line not executable (comment, blank line, etc.)
- `0`: Line executable but not covered
- `> 0`: Line covered (number indicates hit count)

---

## Examples

See the `examples/` directory for complete working examples:
- `examples/go-coverage/` - Go project with coverage upload
- `examples/js-coverage/` - JavaScript/Node.js project with coverage upload

---

## Support

For issues or questions:
- GitHub Issues: [your-repo-url]
- Documentation: [your-docs-url]
- API Docs: `API.md`
