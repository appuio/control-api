kind_marker := $(localenv_dir)/.kind-setup_complete

curl_args ?= --location --fail --silent --show-error

.DEFAULT_TARGET: kind-setup

.PHONY: kind-setup
kind-setup: export KUBECONFIG = $(KIND_KUBECONFIG)
kind-setup: $(kind_marker) $(localenv_dir_created) ## Creates the kind cluster

.PHONY: kind-clean
kind-clean: export KUBECONFIG = $(KIND_KUBECONFIG)
kind-clean: ## Remove the kind Cluster
	@$(KIND) delete cluster --name $(KIND_CLUSTER) || true
	@rm $(kind_marker) $(KIND_KUBECONFIG) || true

###
### Artifacts
###

$(KIND_KUBECONFIG): export KUBECONFIG = $(KIND_KUBECONFIG)
$(KIND_KUBECONFIG):
	$(localenv_dir)/setup-kind.sh "$(KIND)" "$(KIND_CLUSTER)" "$(KIND_NODE_VERSION)" "$(KIND_KUBECONFIG)"
	@kubectl version
	@kubectl cluster-info

$(kind_marker): export KUBECONFIG = $(KIND_KUBECONFIG)
$(kind_marker): $(KIND_KUBECONFIG)
	@kubectl config use-context kind-$(KIND_CLUSTER)
	@touch $(kind_marker)
