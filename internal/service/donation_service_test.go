package service_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"suporter-backend/internal/config"
	"suporter-backend/internal/domain"
	"suporter-backend/internal/repository"
	"suporter-backend/internal/service"
)

func TestDonationService_CreateDonationValidation(t *testing.T) {
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

	donationRepo := repository.NewDonationRepository(gormDB)
	userRepo := repository.NewUserRepository(gormDB)
	projectRepo := repository.NewProjectRepository(gormDB)
	sseBroker := service.NewSSEBroker()
	cfg := &config.Config{Port: "8080"}

	donationServ := service.NewDonationService(donationRepo, userRepo, projectRepo, sseBroker, cfg)
	ctx := context.Background()

	// 1. Amount too low (< 5000)
	_, err = donationServ.CreateDonation(ctx, domain.CreateDonationRequest{
		StreamerUsername: "streamer1",
		Amount:           1000,
	}, false)
	if err == nil {
		t.Errorf("expected error for amount < 5000, got nil")
	}

	// 2. Streamer not found
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("streamer1", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = donationServ.CreateDonation(ctx, domain.CreateDonationRequest{
		StreamerUsername: "streamer1",
		Amount:           50000,
	}, false)
	if err == nil {
		t.Errorf("expected error for unknown streamer, got nil")
	}

	// 3. Success case
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("streamer1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "role", "is_active", "qris_url"}).
			AddRow(1, "streamer1", "streamer", true, "https://qris.link/qr.png"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "donations" WHERE streamer_id = $1 AND unique_code = $2 AND status = $3`)).
		WithArgs(uint64(1), sqlmock.AnyArg(), "pending").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "donations"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(101))
	mock.ExpectCommit()

	don, err := donationServ.CreateDonation(ctx, domain.CreateDonationRequest{
		StreamerUsername: "streamer1",
		SenderName:       "Budi",
		Amount:           50000,
		Message:          "Semangat live-nya!",
	}, false)
	if err != nil {
		t.Fatalf("unexpected error creating donation: %v", err)
	}

	if don.StreamerID != 1 {
		t.Errorf("expected streamer ID 1, got %d", don.StreamerID)
	}

	if don.Status != "pending" {
		t.Errorf("expected donation status 'pending', got %s", don.Status)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
