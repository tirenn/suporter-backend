package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	"suporter-backend/internal/config"
	"suporter-backend/internal/domain"
	"suporter-backend/internal/repository"
)

type DonationService interface {
	CreateDonation(ctx context.Context, req domain.CreateDonationRequest, isTest bool) (*domain.Donation, error)
	VerifyWebhookDonation(ctx context.Context, webhookKey string, incomingAmount int64, rawBody []byte, timestampHeader, signatureHeader string) (*domain.Donation, error)
}

type donationService struct {
	donationRepo repository.DonationRepository
	userRepo     repository.UserRepository
	projectRepo  repository.ProjectRepository
	sseBroker    *SSEBroker
	cfg          *config.Config
}

func NewDonationService(
	donationRepo repository.DonationRepository,
	userRepo repository.UserRepository,
	projectRepo repository.ProjectRepository,
	sseBroker *SSEBroker,
	cfg *config.Config,
) DonationService {
	return &donationService{
		donationRepo: donationRepo,
		userRepo:     userRepo,
		projectRepo:  projectRepo,
		sseBroker:    sseBroker,
		cfg:          cfg,
	}
}

func (s *donationService) CreateDonation(ctx context.Context, req domain.CreateDonationRequest, isTest bool) (*domain.Donation, error) {
	if req.Amount < 5000 || req.Amount > 10000000 {
		return nil, fmt.Errorf("donation amount must be between Rp 5.000 and Rp 10.000.000")
	}

	streamer, err := s.userRepo.FindByUsername(ctx, req.StreamerUsername)
	if err != nil {
		return nil, fmt.Errorf("streamer not found")
	}

	if streamer.Role != "streamer" {
		return nil, fmt.Errorf("recipient user is not a streamer")
	}

	if !streamer.IsActive {
		return nil, fmt.Errorf("streamer sedang tidak aktif")
	}

	// Generate 3-digit unique code (100 - 999) checking if there is a pending donation with the same code for this streamer
	var uniqueCode int
	maxAttempts := 900 // Number of possible 3-digit codes
	success := false

	for i := 0; i < maxAttempts; i++ {
		nBig, err := rand.Int(rand.Reader, big.NewInt(900))
		code := 100
		if err == nil {
			code = int(nBig.Int64()) + 100
		} else {
			code = 100 + (i % 900)
		}

		exists, err := s.donationRepo.IsUniqueCodePending(ctx, streamer.ID, code)
		if err == nil && !exists {
			uniqueCode = code
			success = true
			break
		}
	}

	if !success {
		return nil, fmt.Errorf("all unique payment identifier codes are currently in use for this streamer. Please try again later")
	}

	totalAmount := req.Amount + int64(uniqueCode)

	donation := &domain.Donation{
		StreamerID:  streamer.ID,
		SenderName:  req.SenderName,
		Amount:      req.Amount,
		UniqueCode:  uniqueCode,
		TotalAmount: totalAmount,
		Message:     req.Message,
		Status:      "pending",
		IsTest:      isTest,
	}

	if err := s.donationRepo.Create(ctx, donation); err != nil {
		return nil, err
	}

	return donation, nil
}

func (s *donationService) VerifyWebhookDonation(ctx context.Context, webhookKey string, incomingAmount int64, rawBody []byte, timestampHeader, signatureHeader string) (*domain.Donation, error) {
	streamer, err := s.userRepo.FindByWebhookKey(ctx, webhookKey)
	if err != nil {
		return nil, fmt.Errorf("unauthorized: invalid webhook key")
	}

	// If HMAC verification is enabled on backend via .env
	if s.cfg != nil && s.cfg.EnableWebhookHMAC {
		if timestampHeader == "" || signatureHeader == "" {
			return nil, fmt.Errorf("unauthorized: missing X-Suporter-Timestamp or X-Suporter-Signature header")
		}

		tsInt, err := strconv.ParseInt(timestampHeader, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unauthorized: invalid timestamp format")
		}

		// 5-minute replay tolerance window (300 seconds)
		now := time.Now().Unix()
		if math.Abs(float64(now-tsInt)) > 300 {
			return nil, fmt.Errorf("unauthorized: webhook request timestamp expired (tolerance window: 5 minutes)")
		}

		// Recompute expected HMAC-SHA256: HMAC(secret, timestamp + "." + rawBody)
		mac := hmac.New(sha256.New, []byte(streamer.WebhookSecret))
		mac.Write([]byte(timestampHeader + "." + string(rawBody)))
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		if subtle.ConstantTimeCompare([]byte(expectedSignature), []byte(signatureHeader)) != 1 {
			return nil, fmt.Errorf("unauthorized: invalid HMAC signature")
		}
	}

	donation, err := s.donationRepo.FindPendingByTotalAmount(ctx, streamer.ID, incomingAmount)
	if err != nil {
		return nil, fmt.Errorf("no matching pending donation found for amount Rp %d", incomingAmount)
	}

	donation.Status = "completed"
	if err := s.donationRepo.Update(ctx, donation); err != nil {
		return nil, err
	}

	// Find first active project of the streamer to trigger OBS overlay alert
	projects, err := s.projectRepo.FindByUserID(ctx, streamer.ID)
	if err == nil && len(projects) > 0 {
		targetProject := projects[0]

		formattedAmount := fmt.Sprintf("Rp %d", donation.TotalAmount)

		alertPayload := map[string]string{
			"name":    donation.SenderName,
			"amount":  formattedAmount,
			"message": donation.Message,
		}

		alert := domain.Alert{
			Name:         donation.SenderName,
			Amount:       formattedAmount,
			Message:      donation.Message,
			Type:         "donation",
			Duration:     targetProject.Duration,
			Timestamp:    time.Now().UnixMilli(),
			HTMLTemplate: targetProject.HTMLTemplate,
			CSSStyle:     targetProject.CSSStyle,
			Payload:      alertPayload,
		}

		// Broadcast alert to project overlay
		s.sseBroker.Broadcast(targetProject.UUID.String(), alert)
	}

	return donation, nil
}
