# Makefile for go_proxy

SHELL := /bin/bash

# -----------------------------------------------------------------------------
# COLOR VARIABLES
# -----------------------------------------------------------------------------
CYAN     := \033[36m
GREEN    := \033[32m
BLUE     := \033[34m
YELLOW   := \033[33m
RED      := \033[31m
NO_COLOR := \033[0m

# -----------------------------------------------------------------------------
# DEFAULT GOAL
# -----------------------------------------------------------------------------
.DEFAULT_GOAL := help

# -----------------------------------------------------------------------------
# Project Settings
# -----------------------------------------------------------------------------
BINARY_SERVER := prx-server
BINARY_CLI    := prx
PROTO_DIR     := proto
PROTO_FILE    := $(PROTO_DIR)/reverse.proto
DOCKERFILE    := go_proxy.dockerfile
IMAGE         := ghcr.io/typeterrors/go_proxy
TAG           := $(shell git rev-parse --short HEAD)
HELM_RELEASE  := go-proxy
HELM_CHART    := charts/go-proxy
NAMESPACE     ?= default

# -----------------------------------------------------------------------------
# HELP: List commands
# -----------------------------------------------------------------------------
.PHONY: help
help:
	@echo -e ""
	@echo -e "${CYAN}go_proxy Makefile Commands:${NO_COLOR}"
	@echo -e "  ${GREEN}install${NO_COLOR}       - Ensure Go, protoc, and @gum are installed."
	@echo -e "  ${GREEN}proto${NO_COLOR}         - Generate Go stubs from reverse.proto."
	@echo -e "  ${GREEN}build${NO_COLOR}         - Build server and CLI binaries."
	@echo -e "  ${GREEN}build-all${NO_COLOR}     - Cross-compile for Linux & Windows."
	@echo -e "  ${GREEN}run-server${NO_COLOR}    - Run the server (HTTP + gRPC) locally."
	@echo -e "  ${GREEN}run-cli${NO_COLOR}       - Run the CLI client locally."
	@echo -e "  ${GREEN}fmt${NO_COLOR}           - Format Go code."
	@echo -e "  ${GREEN}vet${NO_COLOR}           - Vet Go code."
	@echo -e "  ${GREEN}lint${NO_COLOR}          - Run golangci-lint."
	@echo -e "  ${GREEN}test${NO_COLOR}          - Run unit tests."
	@echo -e "  ${GREEN}docker-build${NO_COLOR}  - Build Docker image."
	@echo -e "  ${GREEN}docker-push${NO_COLOR}   - Push Docker image."
	@echo -e "  ${GREEN}helm-deploy${NO_COLOR}   - Deploy/upgrade with Helm."
	@echo -e "  ${GREEN}clean${NO_COLOR}         - Remove build artifacts."
	@echo -e ""

# -----------------------------------------------------------------------------
# INSTALL: prerequisites
# -----------------------------------------------------------------------------
.PHONY: install
install:
	@gum style --foreground 212 'Checking prerequisites...'
	# Go
	if ! command -v go &> /dev/null; then \
		@echo -e "${RED}Go not found!${NO_COLOR} Please install from https://golang.org/dl/"; exit 1; \
	else \
		@gum style --foreground 35 'Go is installed'; \
	fi
	# protoc
	if ! command -v protoc &> /dev/null; then \
		@echo -e "${RED}protoc not found!${NO_COLOR} Install protoc for your OS"; exit 1; \
	else \
		@gum style --foreground 35 'protoc is installed'; \
	fi
	# @gum
	if ! command -v @gum &> /dev/null; then \
		@gum style --foreground 202 '@gum not found! Installing...'; \
		@brew install charmbracelet/@gum/@gum || true; \
	else \
		@gum style --foreground 35 '@gum is installed'; \
	fi
	@gum style --foreground 10 'All prerequisites satisfied.'

# -----------------------------------------------------------------------------
# PROTO: generate stubs
# -----------------------------------------------------------------------------
.PHONY: proto
proto:
	@gum style --border normal --border-foreground 220 'Generating gRPC code'
	@protoc --go_out=. --go-grpc_out=. $(PROTO_FILE)
	@gum style --foreground 10 'Proto stubs generated.'

# -----------------------------------------------------------------------------
# BUILD: compile binaries
# -----------------------------------------------------------------------------
.PHONY: build
build: proto
	@gum style --border normal --border-foreground 220 'Building binaries'
	@go build -o $(BINARY_SERVER) ./cmd/server/main.go
	@gum style --foreground 57 "Built server: $(BINARY_SERVER)"
	@go build -o $(BINARY_CLI)    ./cmd/client/client.go
	@gum style --foreground 57 "Built CLI:    $(BINARY_CLI)"
	@gum style --foreground 10 'Build complete.'

# -----------------------------------------------------------------------------
# BUILD-ALL: cross-compile for Linux & Windows
# -----------------------------------------------------------------------------
.PHONY: build-all
build-all: build-linux build-windows

.PHONY: build-linux
build-linux:
	@gum style --border normal --border-foreground 220 'Cross-building Linux'
	@GOOS=linux GOARCH=amd64 go build -o $(BINARY_SERVER)-linux ./cmd/server/main.go
	@GOOS=linux GOARCH=amd64 go build -o $(BINARY_CLI)-linux    ./cmd/client/client.go
	@gum style --foreground 10 'Linux builds ready.'

.PHONY: build-windows
build-windows:
	@gum style --border normal --border-foreground 220 'Cross-building Windows'
	@GOOS=windows GOARCH=amd64 go build -o $(BINARY_SERVER)-win.exe ./cmd/server/main.go
	@GOOS=windows GOARCH=amd64 go build -o $(BINARY_CLI)-win.exe    ./cmd/client/client.go
	@gum style --foreground 10 'Windows builds ready.'

# -----------------------------------------------------------------------------
# RUN: local execution
# -----------------------------------------------------------------------------
.PHONY: run-server
run-server:
	@gum style --foreground 214 'Running server (HTTP+gRPC)...'
	@go run ./cmd/server/main.go

.PHONY: run-cli
run-cli:
	@gum style --foreground 214 'Launching CLI...'
	@go run ./cmd/client/client.go

# -----------------------------------------------------------------------------
# CODE QUALITY
# -----------------------------------------------------------------------------
.PHONY: fmt
fmt:
	@gum style --foreground 33 'Formatting code...'
	@go fmt ./...

.PHONY: vet
vet:
	@gum style --foreground 33 'Running go vet...'
	@go vet ./...

.PHONY: lint
lint:
	@if command -v golangci-lint &> /dev/null; then \
		@golangci-lint run; \
	else \
		@echo -e "${YELLOW}golangci-lint not installed, skipping lint${NO_COLOR}"; \
	fi

# -----------------------------------------------------------------------------
# TEST: unit tests
# -----------------------------------------------------------------------------
# .PHONY: test
# test:
# 	@gum style --foreground 33 'Running tests...'
# 	@go test ./... -timeout 30s

# -----------------------------------------------------------------------------
# DOCKER: build & push
# -----------------------------------------------------------------------------
.PHONY: docker-build
docker-build:
	@gum style --border normal --border-foreground 220 'Building Docker image'
	@docker build -f $(DOCKERFILE) -t $(IMAGE):$(TAG) .
	@gum style --foreground 10 'Docker build done.'

# .PHONY: docker-push
# docker-push:
# 	@gum style --foreground 220 'Pushing Docker image'
# 	@docker push $(IMAGE):$(TAG)
# 	@docker tag $(IMAGE):$(TAG) $(IMAGE):latest
# 	@docker push $(IMAGE):latest
# 	@gum style --foreground 10 'Docker push done.'

# -----------------------------------------------------------------------------
# HELM DEPLOY: release or upgrade
# -----------------------------------------------------------------------------
# .PHONY: helm-deploy
# helm-deploy:
# 	@gum style --border normal --border-foreground 220 'Deploying with Helm'
# 	@helm upgrade --install $(HELM_RELEASE) $(HELM_CHART) \
# 	  --namespace $(NAMESPACE) \
# 	  --set application.image.tag=$(TAG) \
# 	  --set application.JWT_SECRET=$$JWT_SECRET \
# 	  --set global.PRX_KUBE_CONFIG=$$PRX_KUBE_CONFIG
# 	@gum style --foreground 10 'Helm deploy complete.'

# -----------------------------------------------------------------------------
# CLEAN: remove artifacts
# -----------------------------------------------------------------------------
.PHONY: clean
clean:
	@gum style --foreground 196 'Cleaning...'
	@rm -f $(BINARY_SERVER) $(BINARY_CLI)
	@rm -f $(BINARY_SERVER)-* $(BINARY_CLI)-*
	@rm -rf internal/pb/*.go
	@gum style --foreground 10 'Clean complete.'

