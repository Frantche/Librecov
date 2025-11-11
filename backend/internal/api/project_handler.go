package api

import (
	"encoding/json"
	"net/http"

	"github.com/Frantche/Librecov/backend/internal/middleware"
	"github.com/Frantche/Librecov/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProjectHandler handles project-related requests
type ProjectHandler struct {
	db *gorm.DB
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(db *gorm.DB) *ProjectHandler {
	return &ProjectHandler{db: db}
}

// List returns all projects for the current user
func (h *ProjectHandler) List(c *gin.Context) {
	user, _ := middleware.GetCurrentUser(c)

	var projects []models.Project
	query := h.db.Preload("User").Preload("ProjectShares")

	if !user.Admin {
		// Get user's groups
		var userGroups []string
		if user.Groups != "" {
			json.Unmarshal([]byte(user.Groups), &userGroups)
		}

		// Build query to get projects owned by user OR shared with user's groups
		if len(userGroups) > 0 {
			query = query.Where("user_id = ? OR id IN (SELECT project_id FROM project_shares WHERE group_name IN (?) AND deleted_at IS NULL)", user.ID, userGroups)
		} else {
			query = query.Where("user_id = ?", user.ID)
		}
	}

	if err := query.Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// Get returns a single project
func (h *ProjectHandler) Get(c *gin.Context) {
	id := c.Param("id")
	user, _ := middleware.GetCurrentUser(c)

	var project models.Project
	query := h.db.Preload("User").Preload("Builds").Preload("ProjectShares")

	if !user.Admin {
		// Get user's groups
		var userGroups []string
		if user.Groups != "" {
			json.Unmarshal([]byte(user.Groups), &userGroups)
		}

		// Check if user owns project or has access via group
		if len(userGroups) > 0 {
			query = query.Where("user_id = ? OR id IN (SELECT project_id FROM project_shares WHERE group_name IN (?) AND deleted_at IS NULL)", user.ID, userGroups)
		} else {
			query = query.Where("user_id = ?", user.ID)
		}
	}

	if err := query.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project"})
		}
		return
	}

	c.JSON(http.StatusOK, project)
}

// Create creates a new project
func (h *ProjectHandler) Create(c *gin.Context) {
	user, _ := middleware.GetCurrentUser(c)

	var input struct {
		Name          string `json:"name" binding:"required"`
		CurrentBranch string `json:"current_branch"`
		BaseURL       string `json:"base_url"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project := models.Project{
		Name:          input.Name,
		CurrentBranch: input.CurrentBranch,
		BaseURL:       input.BaseURL,
		Token:         generateRandomString(32),
		UserID:        user.ID,
	}

	if err := h.db.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// Update updates a project
func (h *ProjectHandler) Update(c *gin.Context) {
	id := c.Param("id")
	user, _ := middleware.GetCurrentUser(c)

	var project models.Project
	query := h.db

	if !user.Admin {
		query = query.Where("user_id = ?", user.ID)
	}

	if err := query.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project"})
		}
		return
	}

	var input struct {
		Name          string `json:"name"`
		CurrentBranch string `json:"current_branch"`
		BaseURL       string `json:"base_url"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name != "" {
		project.Name = input.Name
	}
	if input.CurrentBranch != "" {
		project.CurrentBranch = input.CurrentBranch
	}
	if input.BaseURL != "" {
		project.BaseURL = input.BaseURL
	}

	if err := h.db.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// Delete deletes a project
func (h *ProjectHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	user, _ := middleware.GetCurrentUser(c)

	var project models.Project
	query := h.db

	if !user.Admin {
		query = query.Where("user_id = ?", user.ID)
	}

	if err := query.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project"})
		}
		return
	}

	if err := h.db.Delete(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted"})
}

// GetShares returns all shares for a project
func (h *ProjectHandler) GetShares(c *gin.Context) {
	projectID := c.Param("id")
	user, _ := middleware.GetCurrentUser(c)

	// Verify user owns the project
	var project models.Project
	query := h.db
	if !user.Admin {
		query = query.Where("user_id = ?", user.ID)
	}

	if err := query.First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project"})
		}
		return
	}

	var shares []models.ProjectShare
	if err := h.db.Where("project_id = ?", projectID).Find(&shares).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shares"})
		return
	}

	c.JSON(http.StatusOK, shares)
}

// CreateShare shares a project with a group
func (h *ProjectHandler) CreateShare(c *gin.Context) {
	projectID := c.Param("id")
	user, _ := middleware.GetCurrentUser(c)

	// Verify user owns the project
	var project models.Project
	query := h.db
	if !user.Admin {
		query = query.Where("user_id = ?", user.ID)
	}

	if err := query.First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project"})
		}
		return
	}

	var input struct {
		GroupName string `json:"group_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify user has the group in their token (unless admin)
	if !user.Admin {
		var userGroups []string
		if user.Groups != "" {
			json.Unmarshal([]byte(user.Groups), &userGroups)
		}

		hasGroup := false
		for _, g := range userGroups {
			if g == input.GroupName {
				hasGroup = true
				break
			}
		}

		if !hasGroup {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this group"})
			return
		}
	}

	// Check if share already exists
	var existingShare models.ProjectShare
	result := h.db.Where("project_id = ? AND group_name = ?", projectID, input.GroupName).First(&existingShare)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Project is already shared with this group"})
		return
	}

	share := models.ProjectShare{
		ProjectID: project.ID,
		GroupName: input.GroupName,
	}

	if err := h.db.Create(&share).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create share"})
		return
	}

	c.JSON(http.StatusCreated, share)
}

// DeleteShare removes a group share from a project
func (h *ProjectHandler) DeleteShare(c *gin.Context) {
	projectID := c.Param("id")
	shareID := c.Param("shareId")
	user, _ := middleware.GetCurrentUser(c)

	// Verify user owns the project
	var project models.Project
	query := h.db
	if !user.Admin {
		query = query.Where("user_id = ?", user.ID)
	}

	if err := query.First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project"})
		}
		return
	}

	var share models.ProjectShare
	if err := h.db.Where("id = ? AND project_id = ?", shareID, projectID).First(&share).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Share not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch share"})
		}
		return
	}

	if err := h.db.Delete(&share).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete share"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Share deleted"})
}

// ListAll returns all projects in the system (admin only)
func (h *ProjectHandler) ListAll(c *gin.Context) {
	var projects []models.Project
	if err := h.db.Preload("User").Preload("ProjectShares").Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}
