kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: ClusterConfiguration
    apiServer:
        extraArgs:
          oidc-issuer-url: ISSUER_KEYCLOAK/auth/realms/REALM
          oidc-client-id: local-dev
          oidc-username-claim: preferred_username
          oidc-username-prefix: "appuio#"
          oidc-groups-claim: groups
