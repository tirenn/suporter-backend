package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"suporter-backend/internal/config"
	"suporter-backend/internal/domain"
	"suporter-backend/internal/repository"
)

const defaultDonationHTML = `<div class="donation-alert-container">
  <div class="donation-header">
    <div class="coin-icon">💰</div>
    <div class="alert-badge">SUPERCHAT DONATION</div>
  </div>
  <div class="donation-amount">{{amount}}</div>
  <div class="donation-donor">{{name}}</div>
  <div class="donation-message-box">
    <p class="donation-message">"{{message}}"</p>
  </div>
</div>`

const defaultDonationCSS = `.donation-alert-container {
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

type ProjectService interface {
	CreateProject(ctx context.Context, userID uint64, req domain.CreateProjectRequest) (*domain.Project, error)
	GetUserProjects(ctx context.Context, userID uint64) ([]*domain.Project, error)
	GetProjectByUUID(ctx context.Context, projectUUID string) (*domain.Project, error)
	UpdateProjectTemplate(ctx context.Context, projectUUID string, req domain.UpdateProjectRequest) (*domain.Project, error)
	DeleteProject(ctx context.Context, projectUUID string) error
}

type projectService struct {
	projectRepo repository.ProjectRepository
	cfg         *config.Config
}

func NewProjectService(projectRepo repository.ProjectRepository, cfg *config.Config) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		cfg:         cfg,
	}
}

func (s *projectService) CreateProject(ctx context.Context, userID uint64, req domain.CreateProjectRequest) (*domain.Project, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	projectUUID := uuid.New()

	htmlTemplate := strings.TrimSpace(req.HTMLTemplate)
	cssStyle := strings.TrimSpace(req.CSSStyle)

	if htmlTemplate == "" {
		htmlTemplate = defaultDonationHTML
	}
	if cssStyle == "" {
		cssStyle = defaultDonationCSS
	}

	duration := req.Duration
	if duration <= 0 {
		duration = 7000
	}

	project := &domain.Project{
		UUID:         projectUUID,
		UserID:       userID,
		Name:         name,
		Description:  strings.TrimSpace(req.Description),
		EventType:    "donation",
		HTMLTemplate: htmlTemplate,
		CSSStyle:     cssStyle,
		Fields:       `["name","amount","message"]`,
		Duration:     duration,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	s.populateOBSUrl(project)
	return project, nil
}

func (s *projectService) GetUserProjects(ctx context.Context, userID uint64) ([]*domain.Project, error) {
	projects, err := s.projectRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, p := range projects {
		s.ensureDefaultTemplate(p)
		s.populateOBSUrl(p)
	}

	return projects, nil
}

func (s *projectService) GetProjectByUUID(ctx context.Context, projectUUID string) (*domain.Project, error) {
	project, err := s.projectRepo.FindByUUID(ctx, projectUUID)
	if err != nil {
		return nil, err
	}

	s.ensureDefaultTemplate(project)
	s.populateOBSUrl(project)
	return project, nil
}

func (s *projectService) UpdateProjectTemplate(ctx context.Context, projectUUID string, req domain.UpdateProjectRequest) (*domain.Project, error) {
	p, err := s.projectRepo.FindByUUID(ctx, projectUUID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		p.Name = strings.TrimSpace(req.Name)
	}
	if req.Description != "" {
		p.Description = strings.TrimSpace(req.Description)
	}
	if req.HTMLTemplate != "" {
		p.HTMLTemplate = strings.TrimSpace(req.HTMLTemplate)
	}
	p.CSSStyle = strings.TrimSpace(req.CSSStyle)

	if req.Duration > 0 {
		p.Duration = req.Duration
	}

	p.EventType = "donation"
	p.Fields = `["name","amount","message"]`

	if err := s.projectRepo.Update(ctx, p); err != nil {
		return nil, err
	}

	s.populateOBSUrl(p)
	return p, nil
}

func (s *projectService) DeleteProject(ctx context.Context, projectUUID string) error {
	return s.projectRepo.Delete(ctx, projectUUID)
}

func (s *projectService) ensureDefaultTemplate(p *domain.Project) {
	if strings.TrimSpace(p.HTMLTemplate) == "" || !strings.Contains(p.HTMLTemplate, "{{amount}}") {
		p.HTMLTemplate = defaultDonationHTML
		p.CSSStyle = defaultDonationCSS
	}
}

func (s *projectService) populateOBSUrl(p *domain.Project) {
	p.OBSUrl = fmt.Sprintf("http://localhost:%s/overlay/%s", s.cfg.Port, p.UUID.String())
}
