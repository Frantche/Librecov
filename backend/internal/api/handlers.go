package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Frantche/Librecov/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BuildHandler handles build-related requests
type BuildHandler struct {
	db *gorm.DB
}

// NewBuildHandler creates a new build handler
func NewBuildHandler(db *gorm.DB) *BuildHandler {
	return &BuildHandler{db: db}
}

// List returns all builds for a project
//
//	@Summary		List builds for a project
//	@Description	Get all builds for a specific project
//	@Tags			builds
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Project ID"
//	@Success		200	{array}		models.Build
//	@Failure		500	{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/api/v1/projects/{id}/builds [get]
func (h *BuildHandler) List(c *gin.Context) {
	projectID := c.Param("id")

	var builds []models.Build
	if err := h.db.Where("project_id = ?", projectID).
		Order("created_at DESC").
		Preload("Project").
		Find(&builds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch builds"})
		return
	}

	c.JSON(http.StatusOK, builds)
}

// Get returns a single build
//
//	@Summary		Get build details
//	@Description	Get detailed information about a specific build
//	@Tags			builds
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Build ID"
//	@Success		200	{object}	models.Build
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/api/v1/builds/{id} [get]
func (h *BuildHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var build models.Build
	if err := h.db.Preload("Project").
		Preload("Jobs").
		Preload("Jobs.Files").
		First(&build, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Build not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch build"})
		}
		return
	}

	c.JSON(http.StatusOK, build)
}

// JobHandler handles job-related requests
type JobHandler struct {
	db *gorm.DB
}

// NewJobHandler creates a new job handler
func NewJobHandler(db *gorm.DB) *JobHandler {
	return &JobHandler{db: db}
}

// Get returns a single job
//
//	@Summary		Get job details
//	@Description	Get detailed information about a specific job
//	@Tags			jobs
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Job ID"
//	@Success		200	{object}	models.Job
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/api/v1/jobs/{id} [get]
func (h *JobHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var job models.Job
	if err := h.db.Preload("Build").
		Preload("Files").
		First(&job, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch job"})
		}
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListByBuild returns all jobs for a build
//
//	@Summary		List jobs for a build
//	@Description	Get all jobs for a specific build
//	@Tags			jobs
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Build ID"
//	@Success		200	{array}		models.Job
//	@Failure		500	{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/api/v1/builds/{id}/jobs [get]
func (h *JobHandler) ListByBuild(c *gin.Context) {
	buildID := c.Param("id")

	var jobs []models.Job
	if err := h.db.Where("build_id = ?", buildID).
		Preload("Build").
		Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch jobs"})
		return
	}

	c.JSON(http.StatusOK, jobs)
}

// CreateJob creates a new job (API endpoint)
func (h *JobHandler) CreateJob(c *gin.Context) {
	var upload CoverallsUpload

	// First try to get JSON from form field "json" (goveralls format)
	jsonStr := c.PostForm("json")
	if jsonStr != "" {
		if err := json.Unmarshal([]byte(jsonStr), &upload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format", "details": err.Error()})
			return
		}
	} else {
		// If no form field, try reading raw request body
		body, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format", "details": "Cannot read request body"})
			return
		}

		if err := json.Unmarshal(body, &upload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format", "details": err.Error()})
			return
		}
	}

	// Find project by token
	var project models.Project
	if err := h.db.Where("token = ?", upload.RepoToken).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid repo token"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Get or create build
	var build models.Build
	commitSHA := ""
	commitMsg := ""
	branch := ""

	if upload.Git != nil {
		commitSHA = upload.Git.Head.ID
		commitMsg = upload.Git.Head.Message
		branch = upload.Git.Branch
	}

	// Get the latest build number for this project
	var maxBuildNum int
	h.db.Model(&models.Build{}).Where("project_id = ?", project.ID).Select("COALESCE(MAX(build_num), 0)").Scan(&maxBuildNum)

	build = models.Build{
		ProjectID: project.ID,
		BuildNum:  maxBuildNum + 1,
		Branch:    branch,
		CommitSHA: commitSHA,
		CommitMsg: commitMsg,
	}

	if err := h.db.Create(&build).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create build"})
		return
	}

	// Create job
	jobNumber := upload.ServiceJobID
	if jobNumber == "" {
		jobNumber = fmt.Sprintf("%d.1", build.BuildNum)
	}

	job := models.Job{
		BuildID:   build.ID,
		JobNumber: jobNumber,
	}

	if err := h.db.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job"})
		return
	}

	// Process source files and calculate coverage
	totalLines := 0
	coveredLines := 0

	for _, sourceFile := range upload.SourceFiles {
		fileLines := 0
		fileCovered := 0

		// Calculate coverage for this file
		for _, cov := range sourceFile.Coverage {
			if cov != nil {
				fileLines++
				if val, ok := cov.(float64); ok && val > 0 {
					fileCovered++
				}
			}
		}

		totalLines += fileLines
		coveredLines += fileCovered

		// Calculate file coverage rate
		var fileCoverageRate float64
		if fileLines > 0 {
			fileCoverageRate = (float64(fileCovered) / float64(fileLines)) * 100
		}

		// Store coverage as JSON string
		coverageJSON, _ := json.Marshal(sourceFile.Coverage)

		jobFile := models.JobFile{
			JobID:        job.ID,
			Name:         sourceFile.Name,
			Source:       sourceFile.Source,
			Coverage:     string(coverageJSON),
			CoverageRate: fileCoverageRate,
		}

		if err := h.db.Create(&jobFile).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job file"})
			return
		}
	}

	// Calculate overall coverage rate
	var coverageRate float64
	if totalLines > 0 {
		coverageRate = (float64(coveredLines) / float64(totalLines)) * 100
	}

	// Update job with coverage rate
	job.CoverageRate = coverageRate
	h.db.Save(&job)

	// Update build with coverage rate
	build.CoverageRate = coverageRate
	h.db.Save(&build)

	// Update project with latest coverage rate
	project.CoverageRate = coverageRate
	h.db.Save(&project)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Coverage uploaded successfully",
		"project_id":    project.ID,
		"build_id":      build.ID,
		"job_id":        job.ID,
		"coverage_rate": coverageRate,
	})
}

// CoverallsUpload represents the Coveralls JSON format
type CoverallsUpload struct {
	RepoToken     string `json:"repo_token" binding:"required"`
	ServiceName   string `json:"service_name"`
	ServiceNumber string `json:"service_number"`
	ServiceJobID  string `json:"service_job_id"`
	Git           *struct {
		Head struct {
			ID      string `json:"id"`
			Message string `json:"message"`
		} `json:"head"`
		Branch string `json:"branch"`
	} `json:"git"`
	SourceFiles []struct {
		Name     string        `json:"name"`
		Source   string        `json:"source"`
		Coverage []interface{} `json:"coverage"`
	} `json:"source_files"`
}

// Upload handles coverage upload (Coveralls-compatible)
//
//	@Summary		Upload coverage data
//	@Description	Upload code coverage data in Coveralls JSON format
//	@Tags			coverage
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CoverallsUpload	true	"Coverage data"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/upload/v2 [post]
func (h *JobHandler) Upload(c *gin.Context) {
	fmt.Printf("DEBUG: Upload called\n")
	var upload CoverallsUpload

	// First try to bind as direct JSON
	if err := c.ShouldBindJSON(&upload); err != nil {
		// If that fails, try to get JSON from form field "json" (goveralls format)
		jsonStr := c.PostForm("json")
		if jsonStr != "" {
			if err := json.Unmarshal([]byte(jsonStr), &upload); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format", "details": err.Error()})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format", "details": err.Error()})
			return
		}
	}

	// Find project by token
	var project models.Project
	if err := h.db.Where("token = ?", upload.RepoToken).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid repo token"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Get or create build
	var build models.Build
	commitSHA := ""
	commitMsg := ""
	branch := ""

	if upload.Git != nil {
		commitSHA = upload.Git.Head.ID
		commitMsg = upload.Git.Head.Message
		branch = upload.Git.Branch
	}

	// Get the latest build number for this project
	var maxBuildNum int
	h.db.Model(&models.Build{}).Where("project_id = ?", project.ID).Select("COALESCE(MAX(build_num), 0)").Scan(&maxBuildNum)

	build = models.Build{
		ProjectID: project.ID,
		BuildNum:  maxBuildNum + 1,
		Branch:    branch,
		CommitSHA: commitSHA,
		CommitMsg: commitMsg,
	}

	if err := h.db.Create(&build).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create build"})
		return
	}

	// Create job
	jobNumber := upload.ServiceJobID
	if jobNumber == "" {
		jobNumber = fmt.Sprintf("%d.1", build.BuildNum)
	}

	job := models.Job{
		BuildID:   build.ID,
		JobNumber: jobNumber,
	}

	if err := h.db.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job"})
		return
	}

	// Process source files and calculate coverage
	totalLines := 0
	coveredLines := 0

	for _, sourceFile := range upload.SourceFiles {
		fileLines := 0
		fileCovered := 0

		// Calculate coverage for this file
		for _, cov := range sourceFile.Coverage {
			if cov != nil {
				fileLines++
				if val, ok := cov.(float64); ok && val > 0 {
					fileCovered++
				}
			}
		}

		totalLines += fileLines
		coveredLines += fileCovered

		// Calculate file coverage rate
		var fileCoverageRate float64
		if fileLines > 0 {
			fileCoverageRate = (float64(fileCovered) / float64(fileLines)) * 100
		}

		// Store coverage as JSON string
		coverageJSON, _ := json.Marshal(sourceFile.Coverage)

		jobFile := models.JobFile{
			JobID:        job.ID,
			Name:         sourceFile.Name,
			Source:       sourceFile.Source,
			Coverage:     string(coverageJSON),
			CoverageRate: fileCoverageRate,
		}

		if err := h.db.Create(&jobFile).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job file"})
			return
		}
	}

	// Calculate overall coverage rate
	var coverageRate float64
	if totalLines > 0 {
		coverageRate = (float64(coveredLines) / float64(totalLines)) * 100
	}

	// Update job with coverage rate
	job.CoverageRate = coverageRate
	h.db.Save(&job)

	// Update build with coverage rate
	build.CoverageRate = coverageRate
	h.db.Save(&build)

	// Update project with latest coverage rate
	project.CoverageRate = coverageRate
	h.db.Save(&project)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Coverage uploaded successfully",
		"project_id":    project.ID,
		"build_id":      build.ID,
		"job_id":        job.ID,
		"coverage_rate": coverageRate,
	})
}

// FileHandler handles file-related requests
type FileHandler struct {
	db *gorm.DB
}

// NewFileHandler creates a new file handler
func NewFileHandler(db *gorm.DB) *FileHandler {
	return &FileHandler{db: db}
}

// List returns all files for a job
func (h *FileHandler) List(c *gin.Context) {
	jobID := c.Param("id")

	var files []models.JobFile
	if err := h.db.Where("job_id = ?", jobID).Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
		return
	}

	c.JSON(http.StatusOK, files)
}

// Get returns a single file
func (h *FileHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var file models.JobFile
	if err := h.db.First(&file, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch file"})
		}
		return
	}

	c.JSON(http.StatusOK, file)
}

// UserHandler handles user management
type UserHandler struct {
	db *gorm.DB
}

// NewUserHandler creates a new user handler
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// List returns all users (admin only)
func (h *UserHandler) List(c *gin.Context) {
	var users []models.User
	if err := h.db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// ListForOwnershipTransfer returns all users for ownership transfer (authenticated users)
func (h *UserHandler) ListForOwnershipTransfer(c *gin.Context) {
	var users []models.User
	if err := h.db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// Get returns a single user (admin only)
func (h *UserHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := h.db.Preload("Projects").First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// Update updates a user (admin only)
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		}
		return
	}

	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Admin *bool  `json:"admin"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Email != "" {
		user.Email = input.Email
	}
	if input.Admin != nil {
		user.Admin = *input.Admin
	}

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Delete deletes a user (admin only) and transfers their projects to the admin
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	// Get the current admin user from context
	adminUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	admin, ok := adminUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
		return
	}

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		}
		return
	}

	// Prevent admin from deleting themselves
	if user.ID == admin.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	// Start a transaction to ensure all operations succeed or fail together
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Transfer ownership of all projects owned by the user to the admin
	if err := tx.Model(&models.Project{}).Where("user_id = ?", user.ID).Update("user_id", admin.ID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to transfer project ownership: %v", err)})
		return
	}

	// Delete all user tokens associated with the user
	if err := tx.Where("user_id = ?", user.ID).Delete(&models.UserToken{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete user tokens: %v", err)})
		return
	}

	// Delete the user
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete user: %v", err)})
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to commit transaction: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted and projects transferred to admin"})
}

// WebhookHandler handles webhook requests
type WebhookHandler struct {
	db *gorm.DB
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(db *gorm.DB) *WebhookHandler {
	return &WebhookHandler{db: db}
}

// HandleWebhook processes incoming webhooks
func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// BadgeHandler handles badge generation
type BadgeHandler struct {
	db *gorm.DB
}

// NewBadgeHandler creates a new badge handler
func NewBadgeHandler(db *gorm.DB) *BadgeHandler {
	return &BadgeHandler{db: db}
}

// GetBadge generates and returns a coverage badge
func (h *BadgeHandler) GetBadge(c *gin.Context) {
	id := c.Param("id")

	var project models.Project
	if err := h.db.Where("id = ?", id).First(&project).Error; err != nil {
		c.String(http.StatusNotFound, "Project not found")
		return
	}

	// Simple SVG badge
	badge := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="20">
		<rect width="100" height="20" fill="#555"/>
		<rect x="50" width="50" height="20" fill="#4c1"/>
		<text x="25" y="15" fill="#fff" font-family="Arial" font-size="11">coverage</text>
		<text x="75" y="15" fill="#fff" font-family="Arial" font-size="11">%.1f%%</text>
	</svg>`

	c.Header("Content-Type", "image/svg+xml")
	c.String(http.StatusOK, badge, project.CoverageRate)
}
