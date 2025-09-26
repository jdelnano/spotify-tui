# Makefile for spotify-tui

# Binary name
BINARY_NAME=spotify-tui

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

# Main package path
MAIN_PATH=./cmd/main.go

.PHONY: run build clean test help

## run: Run the application
run:
	$(GORUN) $(MAIN_PATH)

## build: Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

## clean: Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

## test: Run tests
test:
	$(GOTEST) -v ./...

## help: Show this help message
help:
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
