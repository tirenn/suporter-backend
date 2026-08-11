package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"suporter-backend/internal/service"
)

type OverlayHandler struct {
	sseBroker *service.SSEBroker
	staticDir string
}

func NewOverlayHandler(sseBroker *service.SSEBroker, staticDir string) *OverlayHandler {
	return &OverlayHandler{
		sseBroker: sseBroker,
		staticDir: staticDir,
	}
}

func (h *OverlayHandler) ServeOverlay(c *gin.Context) {
	projectUUID := c.Param("uuid")
	if projectUUID == "" {
		c.String(http.StatusBadRequest, "Project UUID required. Usage: /overlay/:uuid")
		return
	}

	overlayFilePath := filepath.Join(h.staticDir, "overlay.html")
	if _, err := os.Stat(overlayFilePath); os.IsNotExist(err) {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Overlay template not found at %s", overlayFilePath))
		return
	}

	c.File(overlayFilePath)
}

func (h *OverlayHandler) ServeSSEStream(c *gin.Context) {
	projectUUID := c.Param("uuid")
	if projectUUID == "" {
		c.String(http.StatusBadRequest, "Project UUID required")
		return
	}

	h.sseBroker.ServeHTTP(c.Writer, c.Request, projectUUID)
}
