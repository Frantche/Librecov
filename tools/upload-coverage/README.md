# Coverage Upload Tool

A standalone Go tool to upload code coverage from Go projects to Librecov.

## Features

- ✅ Parses Go coverage profiles (`coverage.out`)
- ✅ Automatically detects Go module path
- ✅ Extracts git information (branch, commit, message)
- ✅ Converts to Coveralls JSON format
- ✅ Uploads to Librecov with proper authentication

## Installation

### From Source

```bash
cd tools/upload-coverage
go build -o ../../upload-coverage .
```

The binary will be created at the project root: `./upload-coverage`

### Pre-built Binary

If available, download the pre-built binary for your platform.

## Usage

### Basic Usage

```bash
# Set your project token
export PROJECT_TOKEN="your-project-token-here"

# Generate coverage
go test -covermode=count -coverprofile=coverage.out ./...

# Upload to Librecov
./upload-coverage -coverprofile=coverage.out
```

### With Custom URL

```bash
export PROJECT_TOKEN="your-project-token"
export LIBRECOV_URL="https://librecov.example.com"

./upload-coverage -coverprofile=coverage.out
```

### Command-line Options

```bash
./upload-coverage [options]

Options:
  -token string
        Project token (can also use PROJECT_TOKEN env var)
  -url string
        Librecov URL (default: http://localhost:4000)
        Can also use LIBRECOV_URL env var
  -coverprofile string
        Coverage profile file (default: coverage.out)
```

## Examples

### Simple Project

```bash
#!/bin/bash
set -e

# Run tests with coverage
go test -v -covermode=count -coverprofile=coverage.out ./...

# Upload
export PROJECT_TOKEN="8lSuT_lns9webm4I-Xz9gdDuFhTKWPPc"
./upload-coverage -coverprofile=coverage.out
```

### CI/CD Integration

#### GitHub Actions

```yaml
- name: Run tests with coverage
  run: go test -v -covermode=count -coverprofile=coverage.out ./...

- name: Upload coverage to Librecov
  env:
    PROJECT_TOKEN: ${{ secrets.LIBRECOV_PROJECT_TOKEN }}
    LIBRECOV_URL: ${{ secrets.LIBRECOV_URL }}
  run: |
    curl -L -o upload-coverage https://github.com/your-org/librecov/releases/latest/download/upload-coverage
    chmod +x upload-coverage
    ./upload-coverage -coverprofile=coverage.out
```

#### GitLab CI

```yaml
test:
  script:
    - go test -v -covermode=count -coverprofile=coverage.out ./...
    - ./upload-coverage -coverprofile=coverage.out
  variables:
    PROJECT_TOKEN: $LIBRECOV_PROJECT_TOKEN
    LIBRECOV_URL: $LIBRECOV_URL
```

## Output

The tool provides detailed output:

```
==> Parsing coverage profile...
   Processed 8 files
   Overall coverage: 15.03% (165/1098 lines)
   Branch: feat/example-go-javascript, Commit: ff05c7e6

==> Uploading to http://localhost:4000/upload/v2...
{
  "build_id": 1,
  "coverage_rate": 15.027322404371585,
  "job_id": 1,
  "message": "Coverage uploaded successfully",
  "project_id": 1
}

✓ Coverage uploaded successfully!
```

## Troubleshooting

### Error: "Invalid repo token"

- Verify your `PROJECT_TOKEN` is correct
- Check that the project exists in Librecov
- Ensure you're using a project token (not a user token)

### Error: "upload failed (status 401)"

- Your token is invalid or expired
- Generate a new project token from the Librecov UI

### Error: "could not read [file]"

- The coverage file contains module paths that don't exist in your filesystem
- Make sure you run the upload from the same directory where tests were run
- Ensure your `go.mod` is in the project root

### No files processed

- Check that your coverage file exists and is not empty
- Verify the coverage file format (should start with `mode: count`)
- Try running tests again: `go test -covermode=count -coverprofile=coverage.out ./...`

## How It Works

1. **Parse Coverage**: Reads the Go coverage profile and extracts coverage data per file
2. **Resolve Paths**: Finds the Go module path from `go.mod` and converts module paths to filesystem paths
3. **Build Coverage Array**: Creates a Coveralls-compatible coverage array for each file
4. **Extract Git Info**: Gets current branch, commit SHA, and commit message
5. **Upload**: Sends the coverage data to Librecov's `/upload/v2` endpoint

## Development

### Running Tests

```bash
cd tools/upload-coverage
go test ./...
```

### Building for Multiple Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o upload-coverage-linux-amd64

# macOS
GOOS=darwin GOARCH=amd64 go build -o upload-coverage-darwin-amd64

# Windows
GOOS=windows GOARCH=amd64 go build -o upload-coverage-windows-amd64.exe
```

## See Also

- [COVERAGE_UPLOAD.md](../../COVERAGE_UPLOAD.md) - Complete coverage upload guide
- [API.md](../../API.md) - Librecov API documentation
- [README.md](../../README.md) - Librecov project documentation
