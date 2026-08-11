package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"suporter-backend/internal/domain"
	"suporter-backend/internal/service"
)

type TemplateHandler struct {
	templateService service.TemplateService
	projectService  service.ProjectService
	sseBroker       *service.SSEBroker
}

func NewTemplateHandler(templateService service.TemplateService, projectService service.ProjectService, sseBroker *service.SSEBroker) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
		projectService:  projectService,
		sseBroker:       sseBroker,
	}
}

// CreateTemplate godoc
// @Summary Create custom overlay template
// @Description Add a custom alert template with HTML, CSS, and dynamic variables (e.g. {{name}}, {{amount}}, {{description}}).
// @Tags Templates
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param uuid path string true "Project UUID"
// @Param request body domain.CreateTemplateRequest true "Template Definition"
// @Success 201 {object} domain.Template
// @Failure 400 {object} map[string]string
// @Router /api/v1/projects/{uuid}/templates [post]
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	projectUUID := c.Param("uuid")
	project, err := h.projectService.GetProjectByUUID(c.Request.Context(), projectUUID)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var req domain.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	template, err := h.templateService.CreateTemplate(c.Request.Context(), project.ID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetTemplates godoc
// @Summary List custom overlay templates for a project
// @Description Retrieve custom alert templates configured for a project.
// @Tags Templates
// @Produce json
// @Param uuid path string true "Project UUID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /api/v1/projects/{uuid}/templates [get]
func (h *TemplateHandler) GetTemplates(c *gin.Context) {
	projectUUID := c.Param("uuid")
	project, err := h.projectService.GetProjectByUUID(c.Request.Context(), projectUUID)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	templates, err := h.templateService.GetTemplatesByProjectID(c.Request.Context(), project.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"count":     len(templates),
	})
}

// GetTemplateByID godoc
// @Summary Get custom template details
// @Tags Templates
// @Produce json
// @Param uuid path string true "Project UUID"
// @Param id path int true "Template ID"
// @Success 200 {object} domain.Template
// @Failure 404 {object} map[string]string
// @Router /api/v1/projects/{uuid}/templates/{id} [get]
func (h *TemplateHandler) GetTemplateByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := h.templateService.GetTemplateByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// UpdateTemplate godoc
// @Summary Update custom overlay template
// @Tags Templates
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param uuid path string true "Project UUID"
// @Param id path int true "Template ID"
// @Param request body domain.CreateTemplateRequest true "Updated Template"
// @Success 200 {object} domain.Template
// @Failure 400 {object} map[string]string
// @Router /api/v1/projects/{uuid}/templates/{id} [put]
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var req domain.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	template, err := h.templateService.UpdateTemplate(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// DeleteTemplate godoc
// @Summary Delete custom overlay template
// @Tags Templates
// @Security BearerAuth
// @Produce json
// @Param uuid path string true "Project UUID"
// @Param id path int true "Template ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/projects/{uuid}/templates/{id} [delete]
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	if err := h.templateService.DeleteTemplate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// TriggerCustomAlert godoc
// @Summary Trigger live alert using custom overlay template
// @Description Send dynamic key-value payload (e.g. name, amount, description) to pop up on OBS Studio overlay using custom template HTML/CSS.
// @Tags Templates
// @Accept json
// @Produce json
// @Param uuid path string true "Project UUID"
// @Param id path int true "Template ID"
// @Param request body domain.TriggerCustomAlertRequest true "Dynamic Key-Value Payload"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /api/v1/projects/{uuid}/templates/{id}/trigger [post]
func (h *TemplateHandler) TriggerCustomAlert(c *gin.Context) {
	projectUUID := c.Param("uuid")
	project, err := h.projectService.GetProjectByUUID(c.Request.Context(), projectUUID)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := h.templateService.GetTemplateByID(c.Request.Context(), id)
	if err != nil || template == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	var req domain.TriggerCustomAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload request body: " + err.Error()})
		return
	}

	duration := template.Duration
	if req.Duration > 0 {
		duration = req.Duration
	}

	alert := domain.Alert{
		Type:         string(template.EventType),
		Duration:     duration,
		Timestamp:    time.Now().UnixMilli(),
		HTMLTemplate: template.HTMLTemplate,
		CSSStyle:     template.CSSStyle,
		Payload:      req.Payload,
	}

	h.sseBroker.Broadcast(projectUUID, alert)

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"message":      "Custom alert triggered successfully",
		"project_uuid": projectUUID,
		"template_id":  template.ID,
		"alert":        alert,
	})
}
