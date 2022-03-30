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

webhook-certs/tls.key:
	mkdir -p webhook-certs
	openssl req -x509 -newkey rsa:4096 -nodes -keyout webhook-certs/tls.key -out webhook-certs/tls.crt -days 3650 -subj "/CN=webhook-service.control-api.svc" -addext "subjectAltName = DNS:webhook-service.control-api.svc"

$(KIND_KUBECONFIG): export KUBECONFIG = $(KIND_KUBECONFIG)
$(KIND_KUBECONFIG): webhook-certs/tls.key
	$(localenv_dir)/setup-kind.sh "$(KIND)" "$(KIND_CLUSTER)" "$(KIND_NODE_VERSION)" "$(KIND_KUBECONFIG)"
	@kubectl version
	@kubectl cluster-info

$(kind_marker): export KUBECONFIG = $(KIND_KUBECONFIG)
$(kind_marker): $(KIND_KUBECONFIG)
	@kubectl config use-context kind-$(KIND_CLUSTER)
	@touch $(kind_marker)
