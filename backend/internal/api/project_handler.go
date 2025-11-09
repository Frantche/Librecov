package api

import (
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
	query := h.db.Preload("User")

	if !user.Admin {
		query = query.Where("user_id = ?", user.ID)
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
	query := h.db.Preload("User").Preload("Builds")

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
