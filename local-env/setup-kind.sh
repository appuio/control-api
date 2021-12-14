#!/bin/bash
# vim:sts=2:ts=2:et:sw=2:tw=0

set -euo pipefail

readonly script_dir=$(dirname "$0")
readonly kind_cmd="${1:-kind}"
readonly kind_cluster="${2:-control-api-localenv}"
readonly kind_node_version="${3:-v1.22.1}"
readonly kind_kubeconfig="${4:-"${script_dir}/control-api.kubeconfig"}"

export KUBECONFIG="${kind_kubeconfig}"

keycloak_url=https://id.dev.appuio.cloud

step() {
  echo
  echo -e "$1"
  read -n 1 -s -r -p "Press any key to continue"
  echo
}

check_command() {
  if ! command -v "${1}" >/dev/null 2>&1; then
    step "Install ${2}. Follow the instructions at ${3}"
  fi
}

check_command "kubectl" "kubectl" "https://kubernetes.io/docs/tasks/tools/#kubectl"
check_command "kubectl-oidc_login" "kubectl oidc-login plugin" "Follow the instructions at https://github.com/int128/kubelogin#setup"

echo
read -r -p "Provide the URL of the Keycloak to connect the local environment to (default=${keycloak_url}): " user_url
if [ x"${user_url}" != x"" ]; then
  keycloak_url="${user_url}"
fi

echo
identifier=
while [ x"$identifier" == x"" ]; do
  read -r -p "Provide an identifier for your local-dev Keycloak realm: " identifier
done

realm_name="local-dev-${identifier}"
sed -e "s/REPLACEME/${realm_name}/g" "${script_dir}/templates/realm.json.tpl" > "${script_dir}/realm.json"

step "Navigate to ${keycloak_url} and create a new realm by importing the '$(realpath "${script_dir}/realm.json")' file."

step "Create a user in the new realm, grant it realm role 'admin'.\nMake sure the user has an email configured and 'Email Verified' is set to 'On'."

step "Note: In the next step, a browser window will open where you have to sign in to Keycloak with the user you've created in the previous step".

sed -e "s#ISSUER_KEYCLOAK#${keycloak_url}#; s/REALM/${realm_name}/g" "${script_dir}/templates/kind-oidc.yaml.tpl" > "${script_dir}/.kind-oidc.yaml"
${kind_cmd} create cluster \
  --name "${kind_cluster}" \
  --image "kindest/node:${kind_node_version}" \
  --config="${script_dir}/.kind-oidc.yaml"
rm "${script_dir}/.kind-oidc.yaml"
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: oidc-cluster-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: Group
  name: admin
EOF
kubectl oidc-login setup \
    --oidc-issuer-url="${keycloak_url}/auth/realms/${realm_name}" \
    --oidc-client-id=local-dev >/dev/null 2>&1
kubectl config set-credentials oidc-user \
  --exec-api-version=client.authentication.k8s.io/v1beta1 \
  --exec-command=kubectl \
  --exec-arg=oidc-login \
  --exec-arg=get-token \
  --exec-arg=--oidc-issuer-url="${keycloak_url}/auth/realms/${realm_name}" \
  --exec-arg=--oidc-client-id=local-dev \
  --exec-arg=--oidc-extra-scope="email offline_access profile openid"
kubectl config set-context --current --user=oidc-user
kubectl apply -k "${script_dir}/../config/crd/apiextensions.k8s.io/v1"

step "Setup finished. Set environment variable KUBECONFIG to '$(realpath "${kind_kubeconfig}")' to interact with the local dev cluster"
