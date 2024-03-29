# Set Shell to bash, otherwise some targets fail with dash/zsh etc.
SHELL := /bin/bash

# Disable built-in rules
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-builtin-variables
.SUFFIXES:
.SECONDARY:
.DEFAULT_GOAL := help

PROJECT_ROOT_DIR = .
include Makefile.vars.mk

localenv_make := $(MAKE) -C local-env

.PHONY: help
help: ## Show this help
	@grep -E -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

all: build ## Invokes the build target

.PHONY: test
test: ## Run tests
	go test ./... -coverprofile cover.tmp.out
	cat cover.tmp.out | grep -v "zz_generated.deepcopy.go" > cover.out

.PHONY: build
build: generate fmt vet $(BIN_FILENAME) ## Build manager binary

.PHONY: generate
generate: ## Generate manifests e.g. CRD, RBAC etc.
	go generate ./...
	# Generate code
	go run sigs.k8s.io/controller-tools/cmd/controller-gen object paths="./..."
	# Generate CRDs
	go run sigs.k8s.io/controller-tools/cmd/controller-gen webhook paths="./..." output:crd:artifacts:config=$(CRD_ROOT_DIR)/v1/base crd:crdVersions=v1
	go run sigs.k8s.io/controller-tools/cmd/controller-gen rbac:roleName=control-api-apiserver paths="./apiserver/...;./apis/..." output:artifacts:config=config/rbac/apiserver
	go run sigs.k8s.io/controller-tools/cmd/controller-gen rbac:roleName=control-api-controller paths="./controllers/...;./webhooks/..." output:artifacts:config=config/rbac/controller

.PHONY: crd
crd: generate ## Generate CRD to file
	$(KUSTOMIZE) build $(CRD_ROOT_DIR)/v1 > $(CRD_FILE)

.PHONY: fmt
fmt: ## Run go fmt against code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code
	go vet ./...

.PHONY: lint
lint: fmt vet ## All-in-one linting
	@echo 'Check for uncommitted changes ...'
	git diff --exit-code

.PHONY: build.docker
build.docker: $(BIN_FILENAME) ## Build the docker image
	docker build . \
		--tag $(GHCR_IMG)

clean: ## Cleans up the generated resources
	rm -rf dist/ cover.out $(BIN_FILENAME) || true

.PHONY: run-api
KUBECONFIG ?= ~/.kube/config
BE_STORAGE ?= fake
BE_ODOO_URL ?= https://odoo.example.com
run-api: build ## Starts control api apiserver against the current Kubernetes cluster (based on your local config)
	$(BIN_FILENAME) api --secure-port 9443 --kubeconfig $(KUBECONFIG) --authentication-kubeconfig $(KUBECONFIG) --authorization-kubeconfig $(KUBECONFIG) --cluster-roles appuio-organization-viewer,appuio-organization-admin --username-prefix "appuio#" --billing-entity-storage $(BE_STORAGE) --billing-entity-odoo8-url $(BE_ODOO_URL)

.PHONY: run-controller
run-controller: build ## Starts control api controller against the current Kubernetes cluster (based  on your local config)
	$(localenv_make) webhook-certs/tls.key
	$(BIN_FILENAME) controller --username-prefix "appuio#" --webhook-cert-dir=./local-env/webhook-certs --webhook-port=9444 --zap-log-level debug --billingentity-email-cron-interval "@every 1m"

.PHONY: run-cleanup
run-cleanup: build ## Starts cleanup command
	$(BIN_FILENAME) cleanup --billing-entity-odoo8-url $(BE_ODOO_URL)

.PHONY: local-env
local-env-setup: ## Setup local kind-based dev environment
	$(localenv_make) setup

.PHONY: local-env-clean-setup
local-env-clean: ## Clean the local dev environment
	$(localenv_make) clean-setup

###
### Assets
###

# Build the binary without running generators
.PHONY: $(BIN_FILENAME)
$(BIN_FILENAME): export CGO_ENABLED = 0
$(BIN_FILENAME):
	@echo "GOOS=$$(go env GOOS) GOARCH=$$(go env GOARCH)"
	go build -o $(BIN_FILENAME)
