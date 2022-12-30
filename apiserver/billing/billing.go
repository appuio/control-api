package billing

import (
	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	"github.com/appuio/control-api/apiserver/authwrapper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	genericregistry "k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	restbuilder "sigs.k8s.io/apiserver-runtime/pkg/builder/rest"
	"sigs.k8s.io/apiserver-runtime/pkg/util/loopback"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// New returns a new storage provider with RBAC authentication for BillingEntities
func New(stor authwrapper.StorageScoper) restbuilder.ResourceHandlerProvider {
	return func(s *runtime.Scheme, g genericregistry.RESTOptionsGetter) (rest.Storage, error) {
		c, err := client.NewWithWatch(loopback.GetLoopbackMasterClientConfig(), client.Options{})
		if err != nil {
			return nil, err
		}

		astor, err := authwrapper.NewAuthorizedStorage(stor, metav1.GroupVersionResource{
			Group:    "rbac.appuio.io",
			Version:  "v1",
			Resource: (&billingv1.BillingEntity{}).GetGroupVersionResource().Resource,
		}, loopback.GetAuthorizer())
		if err != nil {
			return nil, err
		}

		stor := &createRBACWrapper{
			storageCreator: astor.(storageCreator),
			client:         c,
		}

		return stor, nil
	}
}
