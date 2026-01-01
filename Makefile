.PHONY: build run watch clean swagger swagger-install test lint docker-build docker-run

# Variables
BINARY_NAME := productivity-timer
MAIN_PACKAGE := ./cmd/api
SERVICE_PORT := 8080
BIN_DIR := ./bin
BIN_PATH := $(BIN_DIR)/$(BINARY_NAME)

## build: Builds the Go application binary.
build: swagger
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_PATH) $(MAIN_PACKAGE)

## run: Builds and runs the application once.
run: build
	$(BIN_PATH)

## watch: Uses air for live reloading during development.
watch:
	mkdir -p $(BIN_DIR)
	air --build.cmd "go build -o $(BIN_PATH) $(MAIN_PACKAGE)" --build.bin "$(BIN_PATH)"

## clean: Removes the built binary and other temporary files.
clean:
	rm -f $(BIN_PATH)
	rm -rf docs/
	rm -f coverage.out
	# Add other cleanup commands here

## swagger-install: Install swag CLI tool for generating Swagger docs.
swagger-install:
	go install github.com/swaggo/swag/cmd/swag@latest

## swagger: Generate Swagger documentation.
swagger:
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

## test: Run all tests with coverage.
test:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

## test-short: Run tests without race detection (faster).
test-short:
	go test -v ./...

## lint: Run golangci-lint.
lint:
	golangci-lint run --timeout=5m

## lint-install: Install golangci-lint.
lint-install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## docker-build: Build the Docker image.
docker-build:
	docker build -t $(BINARY_NAME):latest .

## docker-run: Run the Docker container.
docker-run:
	docker run -p $(SERVICE_PORT):$(SERVICE_PORT) --env-file .env $(BINARY_NAME):latest

## help: Show this help message.
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'