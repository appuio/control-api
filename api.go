package main

import (
	"fmt"
	"os"
	goruntime "runtime"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	genericregistry "k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"
	ctrl "sigs.k8s.io/controller-runtime"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	userv1 "github.com/appuio/control-api/apis/user/v1"
	"github.com/appuio/control-api/apiserver/authwrapper"
	billingStore "github.com/appuio/control-api/apiserver/billing"
	"github.com/appuio/control-api/apiserver/billing/odoostorage"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/countries"
	orgStore "github.com/appuio/control-api/apiserver/organization"
	"github.com/appuio/control-api/apiserver/secretstorage"
	"github.com/appuio/control-api/apiserver/user"
)

// APICommand creates a new command allowing to start the API server
func APICommand() *cobra.Command {
	roles := []string{}
	usernamePrefix := ""
	var allowEmptyBillingEntity bool

	ob := &odooStorageBuilder{}
	ost := orgStore.New(&roles, &usernamePrefix, &allowEmptyBillingEntity)
	ib := &invitationStorageBuilder{usernamePrefix: &usernamePrefix}

	cmd, err := builder.APIServer.
		WithResourceAndHandler(&orgv1.Organization{}, ost).
		WithResourceAndHandler(organizationStatusRegisterer{&orgv1.Organization{}}, ost).
		WithResourceAndHandler(&billingv1.BillingEntity{}, ob.Build).
		WithResourceAndHandler(&userv1.Invitation{}, ib.Build).
		WithResourceAndHandler(secretstorage.NewStatusSubResourceRegisterer(&userv1.Invitation{}), ib.Build).
		WithResourceAndHandler(&userv1.InvitationRedeemRequest{}, ib.BuildRedeem).
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
	cmd.Flags().BoolVar(&allowEmptyBillingEntity, "allow-empty-billing-entity", true, "Allow empty billing entity references")

	cmd.Flags().StringVar(&ob.billingEntityStorage, "billing-entity-storage", "fake", "Storage backend for billing entities. Supported values: fake, odoo8")
	cmd.Flags().BoolVar(&ob.billingEntityFakeMetadataSupport, "billing-entity-fake-metadata-support", false, "Enable metadata support for the fake storage backend")
	cmd.Flags().StringVar(&ob.odoo8URL, "billing-entity-odoo8-url", "http://localhost:8069", "URL of the Odoo instance to use for billing entities")
	cmd.Flags().BoolVar(&ob.odoo8DebugTransport, "billing-entity-odoo8-debug-transport", false, "Enable debug logging for the Odoo transport")
	cmd.Flags().StringVar(&ob.odoo8CountryListPath, "billing-entity-odoo8-country-list", "countries.yaml", "Path to the country list file in the format of [{name: \"Germany\", code: \"DE\", id: 81},...]")

	cmd.Flags().StringVar(&ob.odoo8AccountingContactDisplayName, "billing-entity-odoo8-accounting-contact-display-name", "Accounting", "Display name of the accounting contact")
	cmd.Flags().StringVar(&ob.odoo8LanguagePreference, "billing-entity-odoo8-language-preference", "en_US", "Language preference of the Odoo record")
	cmd.Flags().IntVar(&ob.odoo8PaymentTermID, "billing-entity-odoo8-payment-term-id", 2, "Payment term ID of the Odoo record")

	cmd.Flags().StringVar(&ib.backingNS, "invitation-storage-backing-ns", "default", "Namespace to store invitation secrets in")

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
	billingEntityStorage, odoo8URL, odoo8CountryListPath       string
	odoo8AccountingContactDisplayName, odoo8LanguagePreference string
	odoo8PaymentTermID                                         int
	billingEntityFakeMetadataSupport, odoo8DebugTransport      bool
}

func (o *odooStorageBuilder) Build(s *runtime.Scheme, g genericregistry.RESTOptionsGetter) (rest.Storage, error) {
	switch o.billingEntityStorage {
	case "fake":
		return billingStore.New(odoostorage.NewFakeStorage(o.billingEntityFakeMetadataSupport).(authwrapper.StorageScoper))(s, g)
	case "odoo8":
		countryIDs, err := countries.LoadCountryIDs(o.odoo8CountryListPath)
		if err != nil {
			return nil, err
		}
		return billingStore.New(odoostorage.NewOdoo8Storage(o.odoo8URL, o.odoo8DebugTransport, odoo8.Config{
			AccountingContactDisplayName: o.odoo8AccountingContactDisplayName,
			LanguagePreference:           o.odoo8LanguagePreference,
			PaymentTermID:                o.odoo8PaymentTermID,
			CountryIDs:                   countryIDs,
		}).(authwrapper.StorageScoper))(s, g)
	default:
		return nil, fmt.Errorf("unknown billing entity storage: %s", o.billingEntityStorage)
	}
}

type invitationStorageBuilder struct {
	usernamePrefix *string

	backingNS string
}

func (i *invitationStorageBuilder) Build(s *runtime.Scheme, g genericregistry.RESTOptionsGetter) (rest.Storage, error) {
	return user.NewInvitationStorage(i.backingNS)(s, g)
}

func (i *invitationStorageBuilder) BuildRedeem(s *runtime.Scheme, g genericregistry.RESTOptionsGetter) (rest.Storage, error) {
	return user.NewInvitationRedeemStorage(*i.usernamePrefix)(s, g)
}

type organizationStatusRegisterer struct {
	*orgv1.Organization
}

func (o organizationStatusRegisterer) GetGroupVersionResource() schema.GroupVersionResource {
	gvr := o.Organization.GetGroupVersionResource()
	gvr.Resource = fmt.Sprintf("%s/status", gvr.Resource)
	return gvr
}
