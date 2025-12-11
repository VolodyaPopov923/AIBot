PROJECT_NAME=AIBot
GO_VERSION=1.21
PLAYWRIGHT_VERSION=v0.4000.1

.PHONY: help install build run clean test

help:
	@echo "AIBot - AI Browser Automation Agent"
	@echo ""
	@echo "Available commands:"
	@echo "  make install    - Install dependencies"
	@echo "  make build      - Build the project"
	@echo "  make run        - Run the agent"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean build artifacts"

install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Installing Playwright browsers..."
	go run github.com/playwright-community/playwright-go/cmd/playwright@latest install chromium

build:
	@echo "Building $(PROJECT_NAME)..."
	go build -o bin/aibot ./cmd/agent

run: build
	@echo "Running $(PROJECT_NAME)..."
	./bin/aibot

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Running linter..."
	golangci-lint run ./...
