package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"suporter-backend/internal/domain"
	"suporter-backend/internal/middleware"
	"suporter-backend/internal/service"
)

type ProjectHandler struct {
	projectService service.ProjectService
	sseBroker      *service.SSEBroker
}

func NewProjectHandler(projectService service.ProjectService, sseBroker *service.SSEBroker) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		sseBroker:      sseBroker,
	}
}

// CreateProject godoc
// @Summary Create a new project with donation overlay template
// @Description Create project with name, description, and overlay template HTML/CSS. Generates unique Project UUID v4 & OBS Stream URL.
// @Tags Projects
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.CreateProjectRequest true "Project Info & Template"
// @Success 201 {object} domain.Project
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/v1/projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req domain.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	project, err := h.projectService.CreateProject(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// GetProjects godoc
// @Summary List all projects for authenticated user
// @Description Get user projects with generated OBS Stream URLs and template details.
// @Tags Projects
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/v1/projects [get]
func (h *ProjectHandler) GetProjects(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	projects, err := h.projectService.GetUserProjects(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if projects == nil {
		projects = []*domain.Project{}
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"count":    len(projects),
	})
}

// GetProjectByUUID godoc
// @Summary Get project details by UUID
// @Description Retrieve single project details with OBS Stream URL & template HTML/CSS.
// @Tags Projects
// @Produce json
// @Param uuid path string true "Project UUID"
// @Success 200 {object} domain.Project
// @Failure 404 {object} map[string]string "Project Not Found"
// @Router /api/v1/projects/{uuid} [get]
func (h *ProjectHandler) GetProjectByUUID(c *gin.Context) {
	projectUUID := c.Param("uuid")
	if projectUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project UUID required"})
		return
	}

	project, err := h.projectService.GetProjectByUUID(c.Request.Context(), projectUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject godoc
// @Summary Update project overlay template
// @Description Update project name, description, html_template, css_style, or duration.
// @Tags Projects
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param uuid path string true "Project UUID"
// @Param request body domain.UpdateProjectRequest true "Updated Project Template Info"
// @Success 200 {object} domain.Project
// @Failure 400 {object} map[string]string
// @Router /api/v1/projects/{uuid} [put]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	projectUUID := c.Param("uuid")
	if projectUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project UUID required"})
		return
	}

	var req domain.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	project, err := h.projectService.UpdateProjectTemplate(c.Request.Context(), projectUUID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject godoc
// @Summary Delete project
// @Description Delete project by UUID.
// @Tags Projects
// @Security BearerAuth
// @Produce json
// @Param uuid path string true "Project UUID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/projects/{uuid} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	projectUUID := c.Param("uuid")
	if projectUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project UUID required"})
		return
	}

	if err := h.projectService.DeleteProject(c.Request.Context(), projectUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// TriggerProjectAlert godoc
// @Summary Trigger donation alert to project OBS overlay
// @Description Send required donation variables (name, amount, message) to pop up on OBS Studio overlay in real-time.
// @Tags Projects
// @Accept json
// @Produce json
// @Param uuid path string true "Project UUID"
// @Param request body domain.TriggerAlertRequest true "Donation Alert Payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Project Not Found"
// @Router /api/v1/projects/{uuid}/alert [post]
func (h *ProjectHandler) TriggerProjectAlert(c *gin.Context) {
	projectUUID := c.Param("uuid")
	if projectUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project UUID required"})
		return
	}

	project, err := h.projectService.GetProjectByUUID(c.Request.Context(), projectUUID)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var req domain.TriggerAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: name, amount, and message are required"})
		return
	}

	name := strings.TrimSpace(req.Name)
	amount := strings.TrimSpace(req.Amount)
	message := strings.TrimSpace(req.Message)

	if name == "" || amount == "" || message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, amount, and message are required fields"})
		return
	}

	duration := project.Duration
	if req.Duration > 0 {
		duration = req.Duration
	}

	payloadMap := map[string]string{
		"name":    name,
		"amount":  amount,
		"message": message,
	}

	alert := domain.Alert{
		Name:         name,
		Amount:       amount,
		Message:      message,
		Type:         "donation",
		Duration:     duration,
		Timestamp:    time.Now().UnixMilli(),
		HTMLTemplate: project.HTMLTemplate,
		CSSStyle:     project.CSSStyle,
		Payload:      payloadMap,
	}

	h.sseBroker.Broadcast(projectUUID, alert)

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"message":      "Donation alert triggered successfully for project overlay",
		"project_uuid": projectUUID,
		"alert":        alert,
	})
}
