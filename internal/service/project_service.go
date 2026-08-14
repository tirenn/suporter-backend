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

const defaultDonationHTML = `<div class="cartoon-alert-container">
  <div class="cartoon-header">
    <div class="cartoon-sparkle">💥</div>
    <div class="cartoon-badge">Suporter datang!!!</div>
    <div class="cartoon-sparkle">⚡</div>
  </div>
  <div class="cartoon-hero">
    <span class="cartoon-name">{{name}}</span>
    <span class="cartoon-action">mengirimkan</span>
    <span class="cartoon-amount">Rp {{amount}}</span>
  </div>
  <div class="cartoon-message-bubble">
    <p class="cartoon-message">{{message}}</p>
  </div>
</div>`

const defaultDonationCSS = `@import url('https://fonts.googleapis.com/css2?family=Fredoka:wght@600;700;800&family=Nunito:wght@700;800;900&display=swap');

.cartoon-alert-container {
  background: linear-gradient(135deg, #FFF066 0%, #FFB800 50%, #FF8A00 100%);
  border: 4px solid #1E293B;
  border-radius: 24px;
  padding: 24px 28px;
  max-width: 480px;
  box-shadow: 6px 8px 0px #0F172A, 0 20px 40px rgba(0, 0, 0, 0.25);
  font-family: 'Fredoka', 'Nunito', sans-serif;
  text-align: center;
  position: relative;
  overflow: hidden;
  animation: cartoonPopIn 0.55s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.cartoon-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-bottom: 12px;
}

.cartoon-sparkle {
  font-size: 1.4rem;
  animation: cartoonBounce 0.8s infinite alternate ease-in-out;
}

.cartoon-badge {
  background: #FF4757;
  color: #FFFFFF;
  border: 3px solid #1E293B;
  border-radius: 50px;
  padding: 4px 18px;
  font-size: 1rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  box-shadow: 3px 3px 0px #1E293B;
  transform: rotate(-1deg);
}

.cartoon-hero {
  font-size: 1.35rem;
  font-weight: 800;
  color: #1E293B;
  line-height: 1.35;
  margin-bottom: 14px;
  text-shadow: 1px 1px 0px rgba(255, 255, 255, 0.6);
}

.cartoon-name {
  color: #2E5BFF;
  font-weight: 900;
  text-decoration: underline wavy #FF4757;
  padding: 0 4px;
}

.cartoon-action {
  color: #1E293B;
  font-weight: 700;
  margin: 0 4px;
}

.cartoon-amount {
  color: #059669;
  background: #FFFFFF;
  border: 2.5px solid #1E293B;
  border-radius: 12px;
  padding: 2px 10px;
  font-weight: 900;
  display: inline-block;
  box-shadow: 2px 3px 0px #1E293B;
  margin-left: 4px;
}

.cartoon-message-bubble {
  background: #FFFFFF;
  border: 3.5px solid #1E293B;
  border-radius: 18px;
  padding: 12px 18px;
  box-shadow: 4px 4px 0px #1E293B;
  position: relative;
  margin-top: 6px;
}

.cartoon-message {
  font-family: 'Nunito', sans-serif;
  font-size: 1.05rem;
  font-weight: 800;
  color: #1E293B;
  line-height: 1.4;
  margin: 0;
  word-break: break-word;
}

@keyframes cartoonPopIn {
  0% { transform: scale(0.4) rotate(-8deg); opacity: 0; }
  70% { transform: scale(1.06) rotate(2deg); opacity: 1; }
  100% { transform: scale(1) rotate(0deg); opacity: 1; }
}

@keyframes cartoonBounce {
  from { transform: translateY(0) scale(1); }
  to { transform: translateY(-5px) scale(1.15); }
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
	existing, err := s.projectRepo.FindByUserID(ctx, userID)
	if err == nil && len(existing) > 0 {
		return nil, fmt.Errorf("anda sudah memiliki project overlay aktif. Setiap streamer dibatasi maksimal 1 project")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	projectUUID := uuid.New()

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
		HTMLTemplate: strings.TrimSpace(req.HTMLTemplate),
		CSSStyle:     strings.TrimSpace(req.CSSStyle),
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
