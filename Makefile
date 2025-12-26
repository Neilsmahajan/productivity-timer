.PHONY: build run watch clean

# Variables
BINARY_NAME := productivity-timer
MAIN_PACKAGE := ./cmd/api
SERVICE_PORT := 8080
BIN_DIR := ./bin
BIN_PATH := $(BIN_DIR)/$(BINARY_NAME)

## build: Builds the Go application binary.
build:
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
	# Add other cleanup commands here

## help: Show this help message.
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'