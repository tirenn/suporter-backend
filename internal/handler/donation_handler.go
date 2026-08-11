package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"suporter-backend/internal/domain"
	"suporter-backend/internal/middleware"
	"suporter-backend/internal/service"
)

type DonationHandler struct {
	donationService service.DonationService
}

func NewDonationHandler(donationService service.DonationService) *DonationHandler {
	return &DonationHandler{
		donationService: donationService,
	}
}

// CreateDonation godoc
// @Summary Create a pending QRIS donation (Requires Viewer Role)
// @Description Initiate donation transaction, generate random unique code identifier, calculate final total amount.
// @Tags Donations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.CreateDonationRequest true "Donation parameters"
// @Success 201 {object} domain.Donation
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/v1/donations [post]
func (h *DonationHandler) CreateDonation(c *gin.Context) {
	// Retrieve logged-in Viewer Username
	val, exists := c.Get(middleware.UserUsernameKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user session"})
		return
	}
	username, ok := val.(string)
	if !ok || username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user session"})
		return
	}

	var req domain.CreateDonationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	donation, err := h.donationService.CreateDonation(c.Request.Context(), username, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, donation)
}

// VerifyWebhookDonation godoc
// @Summary Payment Verification Webhook (Requires Webhook Key)
// @Description Callback endpoint to verify payments matching total_amount = amount + unique_code. Triggers overlay alerts for streamers.
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param key query string true "Streamer Webhook Key"
// @Param request body domain.WebhookDonationRequest true "Webhook payment details"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Router /api/v1/webhooks/donation [post]
func (h *DonationHandler) VerifyWebhookDonation(c *gin.Context) {
	webhookKey := strings.TrimSpace(c.Query("key"))
	if webhookKey == "" {
		webhookKey = strings.TrimSpace(c.GetHeader("X-Webhook-Key"))
	}

	if webhookKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing webhook authentication key"})
		return
	}

	var req domain.WebhookDonationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: total amount is required"})
		return
	}

	donation, err := h.donationService.VerifyWebhookDonation(c.Request.Context(), webhookKey, req.Amount)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "Donation payment verified and alert broadcasted successfully",
		"donation": donation,
	})
}
