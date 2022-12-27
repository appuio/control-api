package main

import (
	"os"
	"runtime"

	"sigs.k8s.io/apiserver-runtime/pkg/builder"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/spf13/cobra"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	billingStore "github.com/appuio/control-api/apiserver/billing"
	orgStore "github.com/appuio/control-api/apiserver/organization"
)

// APICommand creates a new command allowing to start the API server
func APICommand() *cobra.Command {
	roles := []string{}
	usernamePrefix := ""
	cmd, err := builder.APIServer.
		WithResourceAndHandler(&orgv1.Organization{}, orgStore.New(&roles, &usernamePrefix)).
		WithResourceAndHandler(&billingv1.BillingEntity{}, billingStore.New()).
		WithoutEtcd().
		ExposeLoopbackAuthorizer().
		ExposeLoopbackMasterClientConfig().
		Build()
	if err != nil {
		ctrl.Log.WithName("setup").Error(err, "Failed to setup API server")
	}
	cmd.Use = "api"
	cmd.Flags().StringSliceVar(&roles, "cluster-roles", []string{}, "Cluster Roles to bind when creating an organization")
	cmd.Flags().StringVar(&usernamePrefix, "username-prefix", "", "Prefix prepended to username claims. Usually the same as \"--oidc-username-prefix\" of the Kubernetes API server")

	rf := cmd.Run
	cmd.Run = func(cmd *cobra.Command, args []string) {
		ctrl.Log.WithName("setup").WithValues(
			"version", version,
			"date", date,
			"commit", commit,
			"go_os", runtime.GOOS,
			"go_arch", runtime.GOARCH,
			"go_version", runtime.Version(),
			"uid", os.Getuid(),
			"gid", os.Getgid(),
		).Info("Starting control-api…")
		rf(cmd, args)
	}

	return cmd
}
