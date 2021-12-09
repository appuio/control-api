# Set Shell to bash, otherwise some targets fail with dash/zsh etc.
SHELL := /bin/bash

# Disable built-in rules
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-builtin-variables
.SUFFIXES:
.SECONDARY:

PROJECT_ROOT_DIR = ..
include ../Makefile.vars.mk
include kind.mk

uname_s := $(shell uname -s)
ifeq ($(uname_s),Linux)
	xargs := xargs --no-run-if-empty
else
	xargs := xargs
endif

.DEFAULT_GOAL := help

.PHONY: setup
setup: export KUBECONFIG = $(KIND_KUBECONFIG)
setup: $(localenv_dir_created) kind-setup ## Setup the local environment

.PHONY: clean-setup
clean-setup: export KUBECONFIG = $(KIND_KUBECONFIG)
clean-setup: kind-clean ## Clean the local environment (e.g. to rerun the setup)
	rm $(localenv_dir_created) || true

.PHONY: help
help: ## Show this help
	@grep -E -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

###
### Artifacts
###

# a marker file must be created, because the date of the
# directory may update when content in it is created/updated,
# which would cause a rebuild / re-initialization of dependants
$(localenv_dir_created):
	@touch $(localenv_dir_created)
