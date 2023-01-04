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


### Running the control-api API server locally

You can run the control-api API server locally against the currently configured Kubernetes cluster with

```bash
make run-api
```

To access the locally running API server, you need to register it with the [kind]-based local environment.
You can do this by applying the following.

The `externalName` needs to be changed to your specific host IP.
When running kind on Linux you can find it with `docker inspect`.

On some docker distributions the host IP is accessible via `host.docker.internal`.

```bash
HOSTIP=$(docker inspect control-api-v1.22.1-control-plane | jq '.[0].NetworkSettings.Networks.kind.Gateway')
# HOSTIP=host.docker.internal # On some docker distributions

cat <<EOF | sed -e "s/172.21.0.1/$HOSTIP/g" | kubectl apply -f -
apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  name: v1.organization.appuio.io
spec:
  insecureSkipTLSVerify: true
  group: organization.appuio.io
  groupPriorityMinimum: 1000
  versionPriority: 15
  service:
    name: apiserver
    namespace: default
    port: 9443
  version: v1
---
apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  name: v1.billing.appuio.io
spec:
  insecureSkipTLSVerify: true
  group: billing.appuio.io
  groupPriorityMinimum: 1000
  versionPriority: 15
  service:
    name: apiserver
    namespace: default
    port: 9443
  version: v1
---
apiVersion: v1
kind: Service
metadata:
  name: apiserver
  namespace: default
spec:
  ports:
  - port: 9443
    protocol: TCP
    targetPort: 9443
  type: ExternalName
  externalName: 172.21.0.1 # Change to host IP
EOF
```

After that you should be able to access your (with `make run` running) API server with

```bash
kubectl get organizations
```

### Running the control-api controller locally

You can run the control-api controller locally against the currently configured Kubernetes cluster with

```bash
make run-controller
```

To access the locally running controller webhook server, you need to register it with the [kind]-based local environment.
You can do this by applying the following manifests:

```
HOSTIP=$(docker inspect control-api-v1.22.1-control-plane | jq '.[0].NetworkSettings.Networks.kind.Gateway')

cat <<EOF | sed -e "s/172.21.0.1/$HOSTIP/g" | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: webhook-service
  namespace: default
spec:
  ports:
  - port: 9444
    protocol: TCP
    targetPort: 9444
  type: ExternalName
  externalName: 172.21.0.1 # Change to host IP
EOF

kubectl patch validatingwebhookconfiguration validating-webhook-configuration \
  -p '{
    "webhooks": [
      {
        "name": "validate-users.appuio.io",
        "clientConfig": {
          "caBundle": "'"$(base64 -w0 "./local-env/webhook-certs/tls.crt)"'",
          "service": {
            "namespace": "default",
            "port": 9444
          }
        }
      }
    ]
  }'
```
