.PHONY: all build build-windows build-linux run test migrate-up migrate-down migrate-create swagger clean

APP_NAME=server
BUILD_DIR=bin
MAIN_FILE=cmd/server/main.go

# OS Detection
ifeq ($(OS),Windows_NT)
	BINARY_EXT=.exe
	RM=powershell -Command "Remove-Item -Force -Recurse -ErrorAction SilentlyContinue"
	SET_LINUX=cmd /c "set GOOS=linux&& set GOARCH=amd64&& go build -o $(BUILD_DIR)/$(APP_NAME)-linux $(MAIN_FILE)"
	SET_WIN=cmd /c "set GOOS=windows&& set GOARCH=amd64&& go build -o $(BUILD_DIR)/$(APP_NAME)-windows.exe $(MAIN_FILE)"
else
	BINARY_EXT=
	RM=rm -rf
	SET_LINUX=GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux $(MAIN_FILE)
	SET_WIN=GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-windows.exe $(MAIN_FILE)
endif

BINARY=$(BUILD_DIR)/$(APP_NAME)$(BINARY_EXT)

all: swagger test build

build:
	@echo "Building backend binary for host OS..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BINARY) $(MAIN_FILE)

build-windows:
	@echo "Cross-compiling backend binary for Windows (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@$(SET_WIN)

build-linux:
	@echo "Cross-compiling backend binary for Linux (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@$(SET_LINUX)

run:
	@echo "Running backend server (loading config from .env)..."
	go run $(MAIN_FILE)

test:
	@echo "Running backend unit tests..."
	go test -v ./...

swagger:
	@echo "Auto-generating Swagger API documentation..."
	@swag init -g cmd/server/main.go -o docs || go run github.com/swaggo/swag/cmd/swag init -g cmd/server/main.go -o docs

migrate-up:
	@echo "Running Goose migrations UP..."
	@go run $(MAIN_FILE) -migrate-up

migrate-down:
	@echo "Running Goose migrations DOWN..."
	@go run $(MAIN_FILE) -migrate-down

migrate-reset:
	@echo "Resetting database and running Goose migrations from scratch..."
	@go run $(MAIN_FILE) -migrate-reset

migrate-create:
	@echo "Creating Goose SQL migration file..."
	@go run $(MAIN_FILE) -migrate-create="$(name)"

clean:
	@echo "Cleaning build artifacts..."
	$(RM) $(BUILD_DIR) docs
