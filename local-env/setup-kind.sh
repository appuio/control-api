#!/bin/bash
# vim:sts=2:ts=2:et:sw=2:tw=0

set -euo pipefail

readonly script_dir=$(dirname "$0")
readonly kind_cmd="${1:-kind}"
readonly kind_cluster="${2:-control-api-localenv}"
readonly kind_node_version="${3:-v1.25.3}"
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
check_command "kubectl-oidc_login" "kubectl oidc-login plugin" "https://github.com/int128/kubelogin#setup"

echo
read -r -p "Provide the URL of the Keycloak to connect the local environment to (default=${keycloak_url}): " user_url
if [ x"${user_url}" != x"" ]; then
  keycloak_url="${user_url}"
fi

echo
identifier=
while [ x"$identifier" == x"" ]; do
  read -r -p "Provide a suffix for your local-dev Keycloak realm (all local-dev realms are prefixed with 'local-dev-'): " identifier
done

realm_name="local-dev-${identifier}"
sed -e "s/REPLACEME/${realm_name}/g" "${script_dir}/templates/realm.json.tpl" > "${script_dir}/realm.json"

echo -e "\033[1mUsing '${realm_name}' as your local-dev Keycloak realm\033[0m"

step "Navigate to ${keycloak_url}/auth/admin/ and create a new realm by importing the '$(realpath "${script_dir}/realm.json")' file."

step "Create a user in the new realm, grant it 'local-dev' client role 'admin'.\nMake sure the user has an email configured and 'Email Verified' is set to 'On'."

echo ""
echo -e "\033[1m================================================================================"
echo "Note: After the cluster is created, a browser window will open where you have to sign in to Keycloak with the user you've created in the previous step."
echo -e "================================================================================\033[0m"
echo ""

base64_no_wrap='base64'
if [[ "$OSTYPE" == "linux"* ]]; then
  base64_no_wrap='base64 --wrap 0'
fi

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
kubectl apply -k "${script_dir}/../config/deployment"
kubectl apply -k "${script_dir}/../config/user-rbac"
kubectl apply -k "${script_dir}/../config/webhook"
kubectl create secret tls -n control-api webhook-service-tls --cert=webhook-certs/tls.crt --key=webhook-certs/tls.key
kubectl -n control-api patch deployment control-api-controller \
  --type=json \
  -p '[
    {
      "op": "add",
      "path": "/spec/template/spec/containers/0/args/-",
      "value": "--webhook-cert-dir=/var/run/webhook-service-tls"
    },
    {
      "op": "add",
      "path": "/spec/template/spec/volumes",
      value: [
        {
          "name": "webhook-service-tls",
          "secret": {
            "secretName": "webhook-service-tls"
          }
        }
      ]
    },
    {
      "op": "add",
      "path": "/spec/template/spec/containers/0/volumeMounts",
      "value": [
        {
          "name": "webhook-service-tls",
          "mountPath": "/var/run/webhook-service-tls",
          "readOnly": true
        }
      ]
    }
  ]'
kubectl patch validatingwebhookconfiguration validating-webhook-configuration \
  -p '{
    "webhooks": [
      {
        "name": "validate-users.appuio.io",
        "clientConfig": {
          "caBundle": "'"$(eval $base64_no_wrap < "${script_dir}"/webhook-certs/tls.crt)"'"
        }
      },
      {
        "name": "validate-invitations.user.appuio.io",
        "clientConfig": {
          "caBundle": "'"$(eval $base64_no_wrap < "${script_dir}"/webhook-certs/tls.crt)"'"
        }
      }
    ]
  }'


echo =======
echo "Setup finished. To interact with the local dev cluster, set the KUBECONFIG environment variable as follows:"
echo "\"export KUBECONFIG=$(realpath "${kind_kubeconfig}")\""
echo =======
