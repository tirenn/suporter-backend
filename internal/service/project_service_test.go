package service_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"suporter-backend/internal/config"
	"suporter-backend/internal/domain"
	"suporter-backend/internal/repository"
	"suporter-backend/internal/service"
)

func TestProjectService_CreateProjectLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %s", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm database: %s", err)
	}

	projectRepo := repository.NewProjectRepository(gormDB)
	cfg := &config.Config{
		Port: "8080",
	}
	projectServ := service.NewProjectService(projectRepo, cfg)

	ctx := context.Background()

	// 1. Success case: user has 0 projects
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "projects" WHERE user_id = $1 ORDER BY created_at desc`)).
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "projects"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	proj, err := projectServ.CreateProject(ctx, 1, domain.CreateProjectRequest{
		Name:        "Main Overlay",
		Description: "OBS Overlay",
	})
	if err != nil {
		t.Fatalf("unexpected error creating project: %v", err)
	}
	if proj.Name != "Main Overlay" {
		t.Errorf("expected project name 'Main Overlay', got %s", proj.Name)
	}

	// 2. Reject case: user already has 1 project
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "projects" WHERE user_id = $1 ORDER BY created_at desc`)).
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "uuid", "user_id", "name"}).
			AddRow(1, uuid.New().String(), 1, "Existing Project"))

	_, err = projectServ.CreateProject(ctx, 1, domain.CreateProjectRequest{
		Name: "Second Overlay",
	})
	if err == nil {
		t.Fatalf("expected error when creating 2nd project, but got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
