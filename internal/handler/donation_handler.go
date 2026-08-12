package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"suporter-backend/internal/domain"
	"suporter-backend/internal/middleware"
	"suporter-backend/internal/service"
)

type DonationHandler struct {
	donationService    service.DonationService
	recaptchaSecretKey string
	webhookSecret      string
}

func NewDonationHandler(donationService service.DonationService, recaptchaSecretKey, webhookSecret string) *DonationHandler {
	return &DonationHandler{
		donationService:    donationService,
		recaptchaSecretKey: recaptchaSecretKey,
		webhookSecret:      webhookSecret,
	}
}

// CreateDonation godoc
// @Summary Create a pending QRIS donation (Public — No Auth Required)
// @Description Initiate donation transaction, generate random unique code identifier, calculate final total amount. Requires Google reCAPTCHA v2 token.
// @Tags Donations
// @Accept json
// @Produce json
// @Param request body domain.CreateDonationRequest true "Donation parameters"
// @Success 201 {object} domain.Donation
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 429 {object} map[string]string "Too Many Requests"
// @Router /api/v1/donations [post]
func (h *DonationHandler) CreateDonation(c *gin.Context) {
	var req domain.CreateDonationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Verify Google reCAPTCHA v2 token
	if err := middleware.VerifyRecaptcha(h.recaptchaSecretKey, req.RecaptchaToken, c.ClientIP()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verifikasi CAPTCHA gagal: " + err.Error()})
		return
	}

	donation, err := h.donationService.CreateDonation(c.Request.Context(), req)
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
// @Param X-Suporter-Key header string true "Streamer Webhook Key"
// @Param X-Suporter-Signature header string true "HMAC SHA256 Signature"
// @Param request body domain.WebhookDonationRequest true "Webhook payment details"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Router /api/v1/webhooks/donation [post]
func (h *DonationHandler) VerifyWebhookDonation(c *gin.Context) {
	webhookKey := strings.TrimSpace(c.GetHeader("X-Suporter-Key"))
	if webhookKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing X-Suporter-Key header"})
		return
	}

	signature := strings.TrimSpace(c.GetHeader("X-Suporter-Signature"))
	if signature == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing X-Suporter-Signature header"})
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	// Restore body for ShouldBindJSON
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Verify HMAC SHA256 Signature
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(bodyBytes)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid HMAC signature"})
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
