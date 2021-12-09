[![Build](https://img.shields.io/github/workflow/status/appuio/control-api/Test)](https://github.com/appuio/control-api/actions?query=workflow%3ATest)
![Go version](https://img.shields.io/github/go-mod/go-version/appuio/control-api)
![Kubernetes version](https://img.shields.io/badge/k8s-v1.22-blue)
[![Version](https://img.shields.io/github/v/release/appuio/control-api)](https://github.com/appuio/control-api/releases)
[![Maintainability](https://img.shields.io/codeclimate/maintainability/appuio/control-api)](https://codeclimate.com/github/appuio/control-api)
[![GitHub downloads](https://img.shields.io/github/downloads/appuio/control-api/total)](https://github.com/appuio/control-api/releases)

# control-api


## Generate Kubernetes code

If you make changes to the CRD structs you'll need to run code generation.
This can be done with make:

```bash
make generate
```

## Building

See `make help` for a list of build targets.

* `make build`: Build binary for linux/amd64
* `make build -e GOOS=darwin -e GOARCH=arm64`: Build binary for macos/arm64
* `make build.docker`: Build Docker image for local environment

## Install CRDs

CRDs can be installed on the cluster by running `kubectl apply -k config/crd/apiextensions.k8s.io/v1`.
