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

type ProjectService interface {
	CreateProject(ctx context.Context, userID uint64, req domain.CreateProjectRequest) (*domain.Project, error)
	GetUserProjects(ctx context.Context, userID uint64) ([]*domain.Project, error)
	GetProjectByUUID(ctx context.Context, projectUUID string) (*domain.Project, error)
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

	project := &domain.Project{
		UUID:        projectUUID,
		UserID:      userID,
		Name:        name,
		Description: strings.TrimSpace(req.Description),
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
		s.populateOBSUrl(p)
	}

	return projects, nil
}

func (s *projectService) GetProjectByUUID(ctx context.Context, projectUUID string) (*domain.Project, error) {
	project, err := s.projectRepo.FindByUUID(ctx, projectUUID)
	if err != nil {
		return nil, err
	}

	s.populateOBSUrl(project)
	return project, nil
}

func (s *projectService) populateOBSUrl(p *domain.Project) {
	p.OBSUrl = fmt.Sprintf("http://localhost:%s/overlay/%s", s.cfg.Port, p.UUID.String())
}
