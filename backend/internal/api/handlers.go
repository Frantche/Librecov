package api

import (
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
func (h *BuildHandler) List(c *gin.Context) {
	projectID := c.Param("projectId")

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
func (h *BuildHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var build models.Build
	if err := h.db.Preload("Project").
		Preload("Jobs").
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
func (h *JobHandler) ListByBuild(c *gin.Context) {
	buildID := c.Param("buildId")

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
	// This will be implemented with coverage upload logic
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// Upload handles coverage upload (Coveralls-compatible)
func (h *JobHandler) Upload(c *gin.Context) {
	// This will be implemented with Coveralls-compatible upload logic
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
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
	jobID := c.Param("jobId")

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

// Delete deletes a user (admin only)
func (h *UserHandler) Delete(c *gin.Context) {
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

	if err := h.db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
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
	if err := h.db.First(&project, id).Error; err != nil {
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
