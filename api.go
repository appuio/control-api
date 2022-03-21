package main

import (
	"os"
	"runtime"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	orgStore "github.com/appuio/control-api/apiserver/organization"

	"sigs.k8s.io/apiserver-runtime/pkg/builder"
)

func SetupAndStartAPI() {
	logger := newLogger("control-api", true)

	logger.WithValues(
		"version", version,
		"date", date,
		"commit", commit,
		"go_os", runtime.GOOS,
		"go_arch", runtime.GOARCH,
		"go_version", runtime.Version(),
		"uid", os.Getuid(),
		"gid", os.Getgid(),
	).Info("Starting control-api…")

	roles := []string{}
	usernamePrefix := ""
	cmd, err := builder.APIServer.
		WithResourceAndHandler(&orgv1.Organization{}, orgStore.New(&roles, &usernamePrefix)).
		WithoutEtcd().
		ExposeLoopbackAuthorizer().
		ExposeLoopbackMasterClientConfig().
		Build()
	if err != nil {
		logger.Error(err, "Failed to setup API server")
	}

	cmd.Flags().StringSliceVar(&roles, "cluster-roles", []string{}, "Cluster Roles to bind when creating an organization")
	cmd.Flags().StringVar(&usernamePrefix, "username-prefix", "", "Prefix prepended to username claims. Usually the same as \"--oidc-username-prefix\" of the Kubernetes API server")
	err = cmd.Execute()
	if err != nil {
		logger.Error(err, "API server stopped unexpectedly")
	}
}
