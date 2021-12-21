package organization

import (
	"context"
	"fmt"
	"time"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/duration"
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
	listNamespaces(ctx context.Context) (*corev1.NamespaceList, error)
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

var _ rest.Lister = &organizationStorage{}

func (s organizationStorage) NewList() runtime.Object {
	return &orgv1.OrganizationList{}
}

func (s *organizationStorage) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	//TODO(glrf) Don't ignore list options

	namespaces, err := s.namepaces.listNamespaces(ctx)
	if err != nil {
		return nil, err
	}

	res := orgv1.OrganizationList{
		ListMeta: namespaces.ListMeta,
		Items:    []orgv1.Organization{},
	}

	for _, n := range namespaces.Items {
		res.Items = append(res.Items, *namespaceToOrg(&n))
	}

	return &res, nil
}

func (s *organizationStorage) ConvertToTable(ctx context.Context, obj runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	var table metav1.Table

	orgs := []orgv1.Organization{}
	if meta.IsListType(obj) {
		orgList, ok := obj.(*orgv1.OrganizationList)
		if !ok {
			return nil, fmt.Errorf("not an organization: %#v", obj)
		}
		orgs = orgList.Items
	} else {
		org, ok := obj.(*orgv1.Organization)
		if !ok {
			return nil, fmt.Errorf("not an organization: %#v", obj)
		}
		orgs = append(orgs, *org)
	}

	for _, org := range orgs {
		table.Rows = append(table.Rows, orgToTableRow(&org))
	}

	if opt, ok := tableOptions.(*metav1.TableOptions); !ok || !opt.NoHeaders {
		desc := metav1.ObjectMeta{}.SwaggerDoc()
		table.ColumnDefinitions = []metav1.TableColumnDefinition{
			{Name: "Name", Type: "string", Format: "name", Description: desc["name"]},
			{Name: "Display Name", Type: "string", Description: "Name of the organization"},
			{Name: "Namespace", Type: "string", Description: "Name of the underlying namespace"},
			{Name: "Age", Type: "date", Description: desc["creationTimestamp"]},
		}
	}
	return &table, nil
}

func orgToTableRow(org *orgv1.Organization) metav1.TableRow {
	nsName := ""
	if org.Annotations != nil {
		nsName = org.Annotations[namespaceKey]
	}

	return metav1.TableRow{
		Cells:  []interface{}{org.GetName(), org.Spec.DisplayName, nsName, duration.HumanDuration(time.Since(org.GetCreationTimestamp().Time))},
		Object: runtime.RawExtension{Object: org},
	}

}
