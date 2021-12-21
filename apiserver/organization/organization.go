package organization

import (
	"context"
	"fmt"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/endpoints/filters"
	genericregistry "k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	restbuilder "sigs.k8s.io/apiserver-runtime/pkg/builder/rest"
)

// New returns a new storage provider for Organizations
func New() restbuilder.ResourceHandlerProvider {
	return func(s *runtime.Scheme, g genericregistry.RESTOptionsGetter) (rest.Storage, error) {
		return &organizationStorage{
			namepaces:           &loopbackNamespaceProvider{},
			namespaceAuthorizer: &loopbackNamespaceAuthorizer{},
		}, nil
	}
}

type organizationStorage struct {
	namepaces           namespaceProvider
	namespaceAuthorizer authorizer.Authorizer
}
type namespaceProvider interface {
	getNamespace(ctx context.Context, name string) (*corev1.Namespace, error)
	createNamespace(ctx context.Context, ns *corev1.Namespace) error
}

func (s organizationStorage) New() runtime.Object {
	return &orgv1.Organization{}
}

var _ rest.Scoper = &organizationStorage{}

func (s *organizationStorage) NamespaceScoped() bool {
	return false
}

var _ rest.Getter = &organizationStorage{}

func (s *organizationStorage) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	res := &orgv1.Organization{}
	a, err := filters.GetAuthorizerAttributes(ctx)
	if err != nil {
		return nil, err
	}
	decision, _, err := s.namespaceAuthorizer.Authorize(ctx, a)
	if err != nil {
		return nil, err
	} else if decision != authorizer.DecisionAllow {
		return nil, kerrors.NewNotFound(res.GetGroupVersionResource().GroupResource(), name)
	}

	nsName := orgNameToNamespaceName(name)
	ns, err := s.namepaces.getNamespace(ctx, nsName)
	if err != nil {
		return nil, err
	}

	// TODO(glrf) Check that this is actually an organization and not a random namespace
	return namespaceToOrg(ns), nil
}

var _ rest.Creater = &organizationStorage{}

func (s *organizationStorage) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	org, ok := obj.(*orgv1.Organization)
	if !ok {
		return nil, fmt.Errorf("not an organization: %#v", obj)
	}

	a, err := filters.GetAuthorizerAttributes(ctx)
	if err != nil {
		return nil, err
	}

	decision, _, err := s.namespaceAuthorizer.Authorize(ctx, a)
	if err != nil {
		return nil, err
	} else if decision != authorizer.DecisionAllow {
		return nil, kerrors.NewNotFound(org.GetGroupVersionResource().GroupResource(), org.Name)
	}

	// Validate Org
	if err := createValidation(ctx, obj); err != nil {
		return nil, err
	}

	ns := orgToNamespace(org)
	if err := s.namepaces.createNamespace(ctx, ns); err != nil {
		return nil, err
	}
	return org, nil
}
