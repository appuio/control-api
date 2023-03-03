# Local development environment

We provide a script and some templates to setup a local test environment based on [kind](https://kind.sigs.k8s.io/).
The templates can be found in directory `templates/`.

## Prerequisites

* `bash`
* `sed`
* `kind`
* `kubectl`
* `kubelogin` as `kubectl-oidc_login`
* `OpenSSL` (LibreSSL might not be able to generate the necessary setup for the local dev environment)

The setup script will provide links to the install guides for `kubectl` and `kubelogin` if no appropriate command is found.

## Installation

The `setup-kind.sh` script will guide you through the setup.
There are some steps that you have to perform manually on a Keycloak instance, which the script prompts you for.
The script defaults to VSHN's APPUiO Dev Keycloak instance, but you can provide an URL pointing to a different instance during the install process.

Since the setup script requires a few arguments, we provide a make target to run the script:

```
make setup
```

## Billing Entities

The control-api serves `BillingEntities` normally from Odoo directly as an aggregated API.
For local development, this is hard to set up, so the API server has an in-memory fake store if Odoo isn't connected.

Thus, to create some fake entries, you can manually create them as normal resources.
There's an example in `config/examples/billing-entity.yaml`.

Note that the data is lost when the control-api Pod is shut down.
When this case happens, delete the `be-####-viewer` cluster role and -binding before recreating the objects.
