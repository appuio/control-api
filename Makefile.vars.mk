IMG_TAG ?= latest

CURDIR ?= $(shell pwd)
BIN_FILENAME ?= $(CURDIR)/$(PROJECT_ROOT_DIR)/control-api

CRD_FILE ?= crd.yaml
CRD_ROOT_DIR ?= config/crd/apiextensions.k8s.io

KUSTOMIZE ?= go run sigs.k8s.io/kustomize/kustomize/v4

# Image URL to use all building/pushing image targets
GHCR_IMG ?= ghcr.io/appuio/control-api:$(IMG_TAG)

# Local dev environment setup
localenv_dir ?= $(CURDIR)/$(PROJECT_ROOT_DIR)/local-env
localenv_dir_created = $(localenv_dir)/.created

# Kind config
KIND_NODE_VERSION ?= v1.25.3
KIND ?= go run sigs.k8s.io/kind
KIND_KUBECONFIG ?= $(localenv_dir)/kind-kubeconfig-$(KIND_NODE_VERSION)
KIND_CLUSTER ?= control-api-$(KIND_NODE_VERSION)
