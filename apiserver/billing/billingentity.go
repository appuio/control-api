package billingentity

import (
	"k8s.io/apimachinery/pkg/runtime"
	genericregistry "k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	restbuilder "sigs.k8s.io/apiserver-runtime/pkg/builder/rest"
	"sigs.k8s.io/apiserver-runtime/pkg/util/loopback"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	"github.com/appuio/control-api/apiserver/authwrapper"
	"github.com/appuio/control-api/apiserver/billing/odoo"
	"github.com/appuio/control-api/apiserver/billing/odoo/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch,resourceNames=extension-apiserver-authentication
// +kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=mutatingwebhookconfigurations;validatingwebhookconfigurations,verbs=get;list;watch
// +kubebuilder:rbac:groups="flowcontrol.apiserver.k8s.io",resources=prioritylevelconfigurations;flowschemas,verbs=get;list;watch

// New returns a new storage provider for Organizations
func New() restbuilder.ResourceHandlerProvider {
	return func(s *runtime.Scheme, g genericregistry.RESTOptionsGetter) (rest.Storage, error) {
		stor := &billingEntityStorage{
			storage: fake.NewFakeOdooStorage(false),
		}
		return authwrapper.NewAuthorizedStorage(stor, metav1.GroupVersionResource{
			Group:    "rbac.appuio.io",
			Version:  "v1",
			Resource: "billingentities",
		}, loopback.GetAuthorizer())
	}
}

type billingEntityStorage struct {
	storage odoo.OdooStorage
}

var _ rest.Storage = &billingEntityStorage{}

func (s billingEntityStorage) New() runtime.Object {
	return &billingv1.BillingEntity{}
}

func (s billingEntityStorage) Destroy() {}

var _ rest.Scoper = &billingEntityStorage{}

func (s *billingEntityStorage) NamespaceScoped() bool {
	return false
}
