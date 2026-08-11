package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"suporter-backend/internal/domain"
	"suporter-backend/internal/repository"
)

type TemplateService interface {
	CreateTemplate(ctx context.Context, projectID uint64, req domain.CreateTemplateRequest) (*domain.Template, error)
	GetTemplatesByProjectID(ctx context.Context, projectID uint64) ([]*domain.Template, error)
	GetTemplateByID(ctx context.Context, id uint64) (*domain.Template, error)
	UpdateTemplate(ctx context.Context, id uint64, req domain.CreateTemplateRequest) (*domain.Template, error)
	DeleteTemplate(ctx context.Context, id uint64) error
	SeedDefaultTemplates(ctx context.Context, projectID uint64) error
}

type templateService struct {
	templateRepo repository.TemplateRepository
}

func NewTemplateService(templateRepo repository.TemplateRepository) TemplateService {
	return &templateService{
		templateRepo: templateRepo,
	}
}

func (s *templateService) CreateTemplate(ctx context.Context, projectID uint64, req domain.CreateTemplateRequest) (*domain.Template, error) {
	name := strings.TrimSpace(req.Name)
	htmlTemplate := strings.TrimSpace(req.HTMLTemplate)

	if name == "" || req.EventType == "" || htmlTemplate == "" {
		return nil, errors.New("name, event_type, and html_template are required")
	}

	eventTypeEnum, err := domain.ParseEventType(req.EventType)
	if err != nil {
		return nil, err
	}

	if req.Duration <= 0 {
		req.Duration = 5000
	}

	fieldsJSON, err := json.Marshal(req.Fields)
	if err != nil {
		fieldsJSON = []byte("[]")
	}

	t := &domain.Template{
		ProjectID:    projectID,
		Name:         name,
		EventType:    eventTypeEnum,
		HTMLTemplate: htmlTemplate,
		CSSStyle:     strings.TrimSpace(req.CSSStyle),
		Fields:       string(fieldsJSON),
		Duration:     req.Duration,
	}

	if err := s.templateRepo.Create(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *templateService) GetTemplatesByProjectID(ctx context.Context, projectID uint64) ([]*domain.Template, error) {
	return s.templateRepo.FindByProjectID(ctx, projectID)
}

func (s *templateService) GetTemplateByID(ctx context.Context, id uint64) (*domain.Template, error) {
	return s.templateRepo.FindByID(ctx, id)
}

func (s *templateService) UpdateTemplate(ctx context.Context, id uint64, req domain.CreateTemplateRequest) (*domain.Template, error) {
	t, err := s.templateRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		t.Name = strings.TrimSpace(req.Name)
	}
	if req.EventType != "" {
		eventTypeEnum, err := domain.ParseEventType(req.EventType)
		if err != nil {
			return nil, err
		}
		t.EventType = eventTypeEnum
	}
	if req.HTMLTemplate != "" {
		t.HTMLTemplate = strings.TrimSpace(req.HTMLTemplate)
	}
	t.CSSStyle = strings.TrimSpace(req.CSSStyle)

	if req.Duration > 0 {
		t.Duration = req.Duration
	}

	if req.Fields != nil {
		fieldsJSON, _ := json.Marshal(req.Fields)
		t.Fields = string(fieldsJSON)
	}

	if err := s.templateRepo.Update(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *templateService) DeleteTemplate(ctx context.Context, id uint64) error {
	return s.templateRepo.Delete(ctx, id)
}

func (s *templateService) SeedDefaultTemplates(ctx context.Context, projectID uint64) error {
	// Example 1: Animated Donation Superchat Alert
	donHTML := `<div class="donation-alert-container">
  <div class="donation-header">
    <div class="coin-icon">💰</div>
    <div class="alert-badge">SUPERCHAT DONATION</div>
  </div>
  <div class="donation-amount">{{amount}}</div>
  <div class="donation-donor">{{name}}</div>
  <div class="donation-message-box">
    <p class="donation-message">"{{description}}"</p>
  </div>
</div>`

	donCSS := `.donation-alert-container {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.95), rgba(5, 150, 105, 0.95));
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 20px;
  padding: 28px;
  max-width: 450px;
  color: #ffffff;
  font-family: 'Outfit', sans-serif;
  text-align: center;
  box-shadow: 0 20px 50px rgba(16, 185, 129, 0.5), inset 0 0 15px rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(16px);
  animation: popIn 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.donation-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-bottom: 12px;
}

.coin-icon {
  font-size: 1.5rem;
  animation: bounce 1s infinite alternate;
}

.alert-badge {
  background: rgba(0, 0, 0, 0.35);
  padding: 4px 14px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  color: #a7f3d0;
  text-transform: uppercase;
}

.donation-amount {
  font-size: 2.8rem;
  font-weight: 800;
  color: #fef08a;
  text-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
  letter-spacing: -0.02em;
  margin-bottom: 4px;
}

.donation-donor {
  font-size: 1.4rem;
  font-weight: 700;
  color: #ffffff;
  margin-bottom: 14px;
}

.donation-message-box {
  background: rgba(0, 0, 0, 0.25);
  border: 1px dashed rgba(255, 255, 255, 0.25);
  border-radius: 12px;
  padding: 12px 16px;
}

.donation-message {
  font-size: 0.98rem;
  font-style: italic;
  line-height: 1.5;
  color: #ecfdf5;
}

@keyframes popIn {
  0% { transform: scale(0.6) translateY(30px); opacity: 0; }
  100% { transform: scale(1) translateY(0); opacity: 1; }
}

@keyframes bounce {
  from { transform: translateY(0); }
  to { transform: translateY(-6px); }
}`

	_, err1 := s.CreateTemplate(ctx, projectID, domain.CreateTemplateRequest{
		Name:         "Donation Alert (Green Gold)",
		EventType:    "donation",
		HTMLTemplate: donHTML,
		CSSStyle:     donCSS,
		Fields:       []string{"name", "amount", "description"},
		Duration:     7000,
	})

	// Example 2: New Subscriber Alert
	subHTML := `<div class="template-card sub-card">
  <div class="template-header">⭐ NEW SUBSCRIBER! ⭐</div>
  <div class="template-title">{{name}}</div>
  <div class="template-desc">{{description}}</div>
</div>`

	subCSS := `.sub-card { background: linear-gradient(135deg, #8b5cf6, #ec4899); color: #ffffff; border-radius: 16px; padding: 24px; text-align: center; box-shadow: 0 10px 30px rgba(139,92,246,0.5); }
.template-header { font-size: 0.9rem; font-weight: 800; text-transform: uppercase; letter-spacing: 0.1em; opacity: 0.9; }
.template-title { font-size: 1.8rem; font-weight: 800; margin: 8px 0; }
.template-desc { font-size: 1rem; opacity: 0.95; }`

	_, err2 := s.CreateTemplate(ctx, projectID, domain.CreateTemplateRequest{
		Name:         "New Subscriber Alert",
		EventType:    "subscriber",
		HTMLTemplate: subHTML,
		CSSStyle:     subCSS,
		Fields:       []string{"name", "description"},
		Duration:     6000,
	})

	if err1 != nil {
		return err1
	}
	return err2
}
