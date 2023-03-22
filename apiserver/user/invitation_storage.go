package user

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	genericregistry "k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	restbuilder "sigs.k8s.io/apiserver-runtime/pkg/builder/rest"
	"sigs.k8s.io/apiserver-runtime/pkg/util/loopback"
	"sigs.k8s.io/controller-runtime/pkg/client"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	"github.com/appuio/control-api/apiserver/authwrapper"
	"github.com/appuio/control-api/apiserver/secretstorage"
)

// NewInvitationRedeemStorage creates a new REST storage for InvitationRedeemRequest objects
func NewInvitationRedeemStorage(usernamePrefix string) restbuilder.ResourceHandlerProvider {
	return func(s *runtime.Scheme, g genericregistry.RESTOptionsGetter) (rest.Storage, error) {
		c, err := buildClient()
		if err != nil {
			return nil, err
		}

		stor := &invitationRedeemer{
			client:         c,
			usernamePrefix: usernamePrefix,
		}

		return stor, nil
	}
}

// NewInvitationStorage returns a new storage provider with RBAC authentication for BillingEntities
func NewInvitationStorage(backingNS string) restbuilder.ResourceHandlerProvider {
	return func(s *runtime.Scheme, g genericregistry.RESTOptionsGetter) (rest.Storage, error) {
		c, err := buildClient()
		if err != nil {
			return nil, err
		}

		stor, err := secretstorage.NewStorage(&userv1.Invitation{}, c, backingNS)
		if err != nil {
			return nil, err
		}

		stor = &rbacCreatorIsOwner{
			ScopedStandardStorage: stor,
			client:                c,
		}

		astor, err := authwrapper.NewAuthorizedStorage(stor, metav1.GroupVersionResource{
			Group:    "rbac.appuio.io",
			Version:  "v1",
			Resource: (&userv1.Invitation{}).GetGroupVersionResource().Resource,
		}, loopback.GetAuthorizer())
		if err != nil {
			return nil, err
		}

		return astor, nil
	}
}

func buildClient() (client.WithWatch, error) {
	c, err := client.NewWithWatch(loopback.GetLoopbackMasterClientConfig(), client.Options{})
	if err != nil {
		return nil, err
	}

	if err = userv1.AddToScheme(c.Scheme()); err != nil {
		return nil, err
	}
	if err = rbacv1.AddToScheme(c.Scheme()); err != nil {
		return nil, err
	}

	return c, nil
}
