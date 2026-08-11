package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "suporter-backend/docs"
	"suporter-backend/internal/config"
	"suporter-backend/internal/database"
	"suporter-backend/internal/handler"
	"suporter-backend/internal/middleware"
	"suporter-backend/internal/repository"
	"suporter-backend/internal/service"
)

// @title Suporter API - OBS Studio Overlay Backend
// @version 1.0
// @description High-performance REST API & real-time OBS Studio Browser Source Overlay Backend built with Golang, Gin, GORM, PostgreSQL, and Goose.
// @termsOfService http://swagger.io/terms/

// @contact.name Suporter Support
// @contact.email support@suporter.local

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token.
func main() {
	migrateUpFlag := flag.Bool("migrate-up", false, "Run database migrations UP and exit")
	migrateDownFlag := flag.Bool("migrate-down", false, "Run database migrations DOWN and exit")
	migrateCreateFlag := flag.String("migrate-create", "", "Create a new Goose SQL migration file (e.g. -migrate-create=add_roles_table)")
	dropDBFlag := flag.Bool("drop-db", false, "Drop all database tables and exit")
	migrateResetFlag := flag.Bool("migrate-reset", false, "Drop all database tables and re-run all Goose migrations from scratch")
	flag.Parse()

	if *migrateCreateFlag != "" {
		log.Printf("[CLI] Creating Goose migration file '%s'...", *migrateCreateFlag)
		if err := database.CreateGooseMigration("migrations", *migrateCreateFlag); err != nil {
			log.Fatalf("Goose migration creation failed: %v", err)
		}
		os.Exit(0)
	}

	cfg := config.Load()

	gormDB, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	stdDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("Failed to retrieve sql.DB: %v", err)
	}

	if *dropDBFlag {
		log.Println("[CLI] Dropping all database tables...")
		if err := database.DropAllTables(stdDB); err != nil {
			log.Fatalf("Drop database tables failed: %v", err)
		}
		os.Exit(0)
	}

	if *migrateResetFlag {
		log.Println("[CLI] Resetting database and re-running all Goose migrations from scratch...")
		if err := database.DropAllTables(stdDB); err != nil {
			log.Fatalf("Drop database tables failed: %v", err)
		}
		if err := database.RunGooseMigrations(stdDB, "migrations"); err != nil {
			log.Fatalf("Goose migration UP failed: %v", err)
		}
		os.Exit(0)
	}

	if *migrateUpFlag {
		log.Println("[CLI] Running Goose migrations UP...")
		if err := database.RunGooseMigrations(stdDB, "migrations"); err != nil {
			log.Fatalf("Goose migration UP failed: %v", err)
		}
		os.Exit(0)
	}

	if *migrateDownFlag {
		log.Println("[CLI] Running Goose migrations DOWN...")
		if err := database.RunGooseDown(stdDB, "migrations"); err != nil {
			log.Fatalf("Goose migration DOWN failed: %v", err)
		}
		os.Exit(0)
	}

	// Repositories
	userRepo := repository.NewUserRepository(gormDB)
	projectRepo := repository.NewProjectRepository(gormDB)

	// Services
	authService := service.NewAuthService(userRepo, cfg)
	projectService := service.NewProjectService(projectRepo, cfg)
	sseBroker := service.NewSSEBroker()

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService, sseBroker)

	staticDir := "./static"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		staticDir = filepath.Join("backend", "static")
	}
	overlayHandler := handler.NewOverlayHandler(sseBroker, staticDir)

	// Gin Engine Setup
	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	// Swagger API Docs Route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public Auth Endpoints
	api := r.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		// Protected Project & Template Routes
		protectedProjects := api.Group("/projects")
		protectedProjects.Use(middleware.AuthMiddleware(authService))
		{
			protectedProjects.POST("", projectHandler.CreateProject)
			protectedProjects.GET("", projectHandler.GetProjects)
			protectedProjects.PUT("/:uuid", projectHandler.UpdateProject)
			protectedProjects.DELETE("/:uuid", projectHandler.DeleteProject)
		}

		// Public Project & Overlay Trigger Route
		api.GET("/projects/:uuid", projectHandler.GetProjectByUUID)
		api.POST("/projects/:uuid/alert", projectHandler.TriggerProjectAlert)
	}

	// OBS Overlay Routes
	r.GET("/overlay/:uuid", overlayHandler.ServeOverlay)
	r.GET("/overlay/:uuid/stream", overlayHandler.ServeSSEStream)

	// Static Assets & Dashboard Route
	r.Static("/static", staticDir)
	r.StaticFile("/overlay.css", filepath.Join(staticDir, "overlay.css"))
	r.StaticFile("/overlay.js", filepath.Join(staticDir, "overlay.js"))

	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(staticDir, "dashboard.html"))
	})
	r.GET("/dashboard", func(c *gin.Context) {
		c.File(filepath.Join(staticDir, "dashboard.html"))
	})

	// Server instance setup
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // 0 for long-lived SSE streaming
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Println("==================================================================")
		log.Println("  🚀 SUPORTER BACKEND (1 Project = 1 OBS Template Architecture)  ")
		log.Println("==================================================================")
		log.Printf(" > Swagger API Docs      : http://localhost:%s/swagger/index.html", cfg.Port)
		log.Printf(" > Dashboard Web UI      : http://localhost:%s/dashboard", cfg.Port)
		log.Printf(" > Register Endpoint     : POST http://localhost:%s/api/v1/auth/register", cfg.Port)
		log.Printf(" > Login Endpoint        : POST http://localhost:%s/api/v1/auth/login", cfg.Port)
		log.Printf(" > Create Project        : POST http://localhost:%s/api/v1/projects", cfg.Port)
		log.Printf(" > Update Project        : PUT http://localhost:%s/api/v1/projects/{uuid}", cfg.Port)
		log.Printf(" > OBS Overlay Pattern   : http://localhost:%s/overlay/{project_uuid}", cfg.Port)
		log.Printf(" > Trigger Alert Pattern : POST http://localhost:%s/api/v1/projects/{project_uuid}/alert", cfg.Port)
		log.Println("==================================================================")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server Listen error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("\n[Server] Shutdown signal received. Closing connections...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[Server Warning] Forced shutdown error: %v", err)
	}

	log.Println("[Server] Backend server stopped cleanly.")
}
