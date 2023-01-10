package organization

import (
	"context"
	"errors"
	"strings"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"
	"github.com/appuio/control-api/apiserver/authwrapper"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
	genericregistry "k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	restbuilder "sigs.k8s.io/apiserver-runtime/pkg/builder/rest"
	"sigs.k8s.io/apiserver-runtime/pkg/util/loopback"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch,resourceNames=extension-apiserver-authentication
// +kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=mutatingwebhookconfigurations;validatingwebhookconfigurations,verbs=get;list;watch
// +kubebuilder:rbac:groups="flowcontrol.apiserver.k8s.io",resources=prioritylevelconfigurations;flowschemas,verbs=get;list;watch

// New returns a new storage provider for Organizations
func New(clusterRoles *[]string, usernamePrefix *string, allowEmptyBillingEntity *bool) restbuilder.ResourceHandlerProvider {
	return func(s *runtime.Scheme, g genericregistry.RESTOptionsGetter) (rest.Storage, error) {
		masterConfig := loopback.GetLoopbackMasterClientConfig()

		c, err := client.NewWithWatch(loopback.GetLoopbackMasterClientConfig(), client.Options{})
		if err != nil {
			return nil, err
		}
		if err := controlv1.AddToScheme(c.Scheme()); err != nil {
			return nil, err
		}
		if err := billingv1.AddToScheme(c.Scheme()); err != nil {
			return nil, err
		}

		stor := &organizationStorage{
			namepaces: &kubeNamespaceProvider{
				Client: c,
			},
			rbac: kubeRoleBindingCreator{
				Client:       c,
				ClusterRoles: *clusterRoles,
			},
			members: kubeMemberProvider{
				Client: c,
			},
			usernamePrefix:          *usernamePrefix,
			impersonator:            impersonatorFromRestconf{masterConfig, client.Options{Scheme: c.Scheme()}},
			allowEmptyBillingEntity: *allowEmptyBillingEntity,
		}

		return authwrapper.NewAuthorizedStorage(stor, metav1.GroupVersionResource{
			Group:    "rbac.appuio.io",
			Version:  "v1",
			Resource: "organizations",
		}, loopback.GetAuthorizer())
	}
}

type organizationStorage struct {
	namepaces namespaceProvider

	members        memberProvider
	usernamePrefix string

	rbac         roleBindingCreator
	clusterRoles []string

	impersonator impersonator

	allowEmptyBillingEntity bool
}

func (s organizationStorage) New() runtime.Object {
	return &orgv1.Organization{}
}

func (s organizationStorage) Destroy() {}

var _ rest.Scoper = &organizationStorage{}
var _ rest.Storage = &organizationStorage{}

func (s *organizationStorage) NamespaceScoped() bool {
	return false
}

func convertNamespaceError(err error) error {
	groupResource := schema.GroupResource{
		Group:    orgv1.GroupVersion.Group,
		Resource: "organizations",
	}
	statusErr := &apierrors.StatusError{}

	if errors.As(err, &statusErr) {
		switch {
		case apierrors.IsNotFound(err):
			return apierrors.NewNotFound(groupResource, statusErr.ErrStatus.Details.Name)
		case apierrors.IsAlreadyExists(err):
			return apierrors.NewAlreadyExists(groupResource, statusErr.ErrStatus.Details.Name)
		}
	}
	return err
}

func userFrom(ctx context.Context, usernamePrefix string) (user.Info, bool) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return nil, false
	}
	if !strings.HasPrefix(user.GetName(), usernamePrefix) {
		return nil, false
	}

	for _, u := range user.GetGroups() {
		if u == "system:serviceaccounts" {
			return nil, false
		}
	}

	return user, true
}
