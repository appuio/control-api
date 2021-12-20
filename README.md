[![Build](https://img.shields.io/github/workflow/status/appuio/control-api/Test)](https://github.com/appuio/control-api/actions?query=workflow%3ATest)
![Go version](https://img.shields.io/github/go-mod/go-version/appuio/control-api)
![Kubernetes version](https://img.shields.io/badge/k8s-v1.23-blue)
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

## Local development environment

You can setup a [kind]-based local environment with

```bash
make local-env-setup
```

See the [local-env/README.md](./local-env/README.md) for more details on the local environment setup.

Please be aware that the productive deployment of the control-api may run on a different Kubernetes distribution than [kind].

[kind]: https://kind.sigs.k8s.io/
