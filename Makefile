APP_NAME := ktailor
IMAGE_RGST ?= localhost:6000
IMAGE_REPO ?= ktailor
IMAGE_TAG ?= latest
IMG := $(IMAGE_RGST)/$(IMAGE_REPO):$(IMAGE_TAG)

MAIN_PATH := ./cmd/$(APP_NAME)
BIN_DIR := ./bin

# Default target when only 'make' is called
.DEFAULT_GOAL := help

.PHONY: build clean test docker-build docker-push deploy rollout help

help: ## Shows this help screen with all available commands
	@echo "========================================"
	@echo "   kTailor - Makefile Overview"
	@echo "========================================"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""

build: ## Builds the local Go binary in ./bin/
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -o $(BIN_DIR)/$(APP_NAME) $(MAIN_PATH)

clean: ## Removes the local ./bin/ directory
	rm -rf $(BIN_DIR)

test: ## Runs the Go tests of the project
	go test ./...

docker-build: ## Builds the Docker image
	docker build -t $(IMG) .

docker-push: ## Pushes the Docker image to the registry
	docker push $(IMG)

deploy: ## Applies RBAC, certs, templates, and manifests to the cluster
	kubectl apply -f deploy/rbac.yaml
	kubectl apply -f deploy/certs.yaml
	kubectl apply -f deploy/templates.yaml
	kubectl apply -f deploy/manifests.yaml

rollout: docker-build docker-push ## Builds, pushes the image, and restarts the kTailor pod
	kubectl rollout restart deployment $(APP_NAME) -n ktailor
