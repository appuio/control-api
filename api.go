package main

import (
	"fmt"
	"os"
	goruntime "runtime"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	genericregistry "k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"
	ctrl "sigs.k8s.io/controller-runtime"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	"github.com/appuio/control-api/apiserver/authwrapper"
	billingStore "github.com/appuio/control-api/apiserver/billing"
	"github.com/appuio/control-api/apiserver/billing/odoostorage"
	orgStore "github.com/appuio/control-api/apiserver/organization"
)

// APICommand creates a new command allowing to start the API server
func APICommand() *cobra.Command {
	roles := []string{}
	usernamePrefix := ""

	ob := &odooStorageBuilder{}

	cmd, err := builder.APIServer.
		WithResourceAndHandler(&orgv1.Organization{}, orgStore.New(&roles, &usernamePrefix)).
		WithResourceAndHandler(&billingv1.BillingEntity{}, ob.Build).
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

	cmd.Flags().StringVar(&ob.billingEntityStorage, "billing-entity-storage", "fake", "Storage backend for billing entities. Supported values: fake, odoo8")
	cmd.Flags().BoolVar(&ob.billingEntityFakeMetadataSupport, "billing-entity-fake-metadata-support", false, "Enable metadata support for the fake storage backend")
	cmd.Flags().StringVar(&ob.odoo8URL, "billing-entity-odoo8-url", "http://localhost:8069", "URL of the Odoo instance to use for billing entities")
	cmd.Flags().BoolVar(&ob.odoo8DebugTransport, "billing-entity-odoo8-debug-transport", false, "Enable debug logging for the Odoo transport")

	rf := cmd.Run
	cmd.Run = func(cmd *cobra.Command, args []string) {
		ctrl.Log.WithName("setup").WithValues(
			"version", version,
			"date", date,
			"commit", commit,
			"go_os", goruntime.GOOS,
			"go_arch", goruntime.GOARCH,
			"go_version", goruntime.Version(),
			"uid", os.Getuid(),
			"gid", os.Getgid(),
		).Info("Starting control-api…")
		rf(cmd, args)
	}

	return cmd
}

type odooStorageBuilder struct {
	billingEntityStorage, odoo8URL                        string
	billingEntityFakeMetadataSupport, odoo8DebugTransport bool
}

func (o *odooStorageBuilder) Build(s *runtime.Scheme, g genericregistry.RESTOptionsGetter) (rest.Storage, error) {
	fmt.Printf("Building storage with options: %#v\n", o)

	switch o.billingEntityStorage {
	case "fake":
		return billingStore.New(odoostorage.NewFakeStorage(o.billingEntityFakeMetadataSupport).(authwrapper.StorageScoper))(s, g)
	case "odoo8":
		return billingStore.New(odoostorage.NewOdoo8Storage(o.odoo8URL, o.odoo8DebugTransport).(authwrapper.StorageScoper))(s, g)
	default:
		return nil, fmt.Errorf("unknown billing entity storage: %s", o.billingEntityStorage)
	}
}
