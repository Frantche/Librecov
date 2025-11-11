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
	"path/filepath"
	"strings"
)

type CoverallsUpload struct {
	RepoToken   string       `json:"repo_token"`
	ServiceName string       `json:"service_name"`
	Git         *GitInfo     `json:"git,omitempty"`
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

type CoverageBlock struct {
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	NumStmt   int
	Count     int
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

	fmt.Println("==> Parsing coverage profile...")
	coverage, err := parseCoverageProfile(*coverProfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing coverage: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("   Processed %d files\n", len(coverage))

	// Calculate summary
	totalLines := 0
	coveredLines := 0
	for _, file := range coverage {
		for _, cov := range file.Coverage {
			if cov != nil {
				totalLines++
				if count, ok := cov.(int); ok && count > 0 {
					coveredLines++
				}
			}
		}
	}
	coverageRate := 0.0
	if totalLines > 0 {
		coverageRate = float64(coveredLines) / float64(totalLines) * 100
	}
	fmt.Printf("   Overall coverage: %.2f%% (%d/%d lines)\n", coverageRate, coveredLines, totalLines)

	// Get git info
	gitInfo := getGitInfo()
	if gitInfo != nil {
		fmt.Printf("   Branch: %s, Commit: %s\n", gitInfo.Branch, gitInfo.Head.ID[:8])
	}

	// Build upload payload
	upload := CoverallsUpload{
		RepoToken:   *token,
		ServiceName: "manual",
		Git:         gitInfo,
		SourceFiles: coverage,
	}

	// Upload
	fmt.Printf("\n==> Uploading to %s...\n", *url)
	if err := uploadCoverage(*url, upload); err != nil {
		fmt.Fprintf(os.Stderr, "Error uploading coverage: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Coverage uploaded successfully!")
}

func parseCoverageProfile(filename string) ([]SourceFile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Get the directory of the coverage file
	coverageDir, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return nil, err
	}
	
	// Find go.mod by walking up the directory tree
	modulePath := ""
	moduleRoot := coverageDir
	for {
		goModPath := filepath.Join(moduleRoot, "go.mod")
		if goModData, err := os.ReadFile(goModPath); err == nil {
			lines := strings.Split(string(goModData), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "module ") {
					modulePath = strings.TrimSpace(strings.TrimPrefix(line, "module"))
					break
				}
			}
			break
		}
		
		parent := filepath.Dir(moduleRoot)
		if parent == moduleRoot {
			break // Reached root
		}
		moduleRoot = parent
	}

	// Parse coverage blocks
	fileBlocks := make(map[string][]CoverageBlock)
	lines := strings.Split(string(data), "\n")

	for _, line := range lines[1:] { // Skip mode line
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
		
		// Convert module path to filesystem path
		if modulePath != "" && strings.HasPrefix(filename, modulePath+"/") {
			relPath := strings.TrimPrefix(filename, modulePath+"/")
			filename = filepath.Join(moduleRoot, relPath)
		}
		
		rangeStr := fileAndRange[1]

		// Parse range: startLine.startCol,endLine.endCol
		rangeParts := strings.Split(rangeStr, ",")
		if len(rangeParts) < 2 {
			continue
		}

		startParts := strings.Split(rangeParts[0], ".")
		endParts := strings.Split(rangeParts[1], ".")

		if len(startParts) < 2 || len(endParts) < 2 {
			continue
		}

		var startLine, startCol, endLine, endCol, numStmt, count int
		fmt.Sscanf(startParts[0], "%d", &startLine)
		fmt.Sscanf(startParts[1], "%d", &startCol)
		fmt.Sscanf(endParts[0], "%d", &endLine)
		fmt.Sscanf(endParts[1], "%d", &endCol)
		fmt.Sscanf(parts[1], "%d", &numStmt)
		fmt.Sscanf(parts[2], "%d", &count)

		fileBlocks[filename] = append(fileBlocks[filename], CoverageBlock{
			StartLine: startLine,
			StartCol:  startCol,
			EndLine:   endLine,
			EndCol:    endCol,
			NumStmt:   numStmt,
			Count:     count,
		})
	}

	// Build source files with coverage arrays
	result := make([]SourceFile, 0, len(fileBlocks))

	for filename, blocks := range fileBlocks {
		// Read source file
		source, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not read %s: %v\n", filename, err)
			continue
		}

		sourceStr := string(source)
		lines := strings.Split(sourceStr, "\n")

		// Initialize coverage array with nulls
		coverage := make([]interface{}, len(lines))

		// Mark covered lines
		for _, block := range blocks {
			for line := block.StartLine; line <= block.EndLine; line++ {
				if line > 0 && line <= len(coverage) {
					// Use 1-indexed to 0-indexed conversion
					idx := line - 1
					if coverage[idx] == nil || block.Count > 0 {
						coverage[idx] = block.Count
					}
				}
			}
		}

		// Get relative path from current directory
		wd, _ := os.Getwd()
		relPath, err := filepath.Rel(wd, filename)
		if err != nil {
			relPath = filename
		}

		result = append(result, SourceFile{
			Name:     relPath,
			Source:   sourceStr,
			Coverage: coverage,
		})
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

	// If URL doesn't contain a path, default to /upload/v2
	uploadURL := baseURL
	if !strings.Contains(baseURL, "/upload") {
		uploadURL = baseURL + "/upload/v2"
	}

	resp, err := http.Post(uploadURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed (status %d): %s", resp.StatusCode, string(body))
	}

	// Pretty print response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err == nil {
		prettyJSON, _ := json.MarshalIndent(response, "", "  ")
		fmt.Printf("%s\n", prettyJSON)
	} else {
		fmt.Printf("%s\n", string(body))
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
