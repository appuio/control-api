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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
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
	getNamespace(ctx context.Context, name string, options *metav1.GetOptions) (*corev1.Namespace, error)
	deleteNamespace(ctx context.Context, name string, options *metav1.DeleteOptions) (*corev1.Namespace, error)
	createNamespace(ctx context.Context, ns *corev1.Namespace, options *metav1.CreateOptions) error
	updateNamespace(ctx context.Context, ns *corev1.Namespace, options *metav1.UpdateOptions) error
	listNamespaces(ctx context.Context, options *metainternalversion.ListOptions) (*corev1.NamespaceList, error)
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
	ns, err := s.namepaces.getNamespace(ctx, nsName, options)
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
	if err := s.namepaces.createNamespace(ctx, ns, options); err != nil {
		return nil, err
	}
	return org, nil
}

var _ rest.Lister = &organizationStorage{}

func (s organizationStorage) NewList() runtime.Object {
	return &orgv1.OrganizationList{}
}

func (s *organizationStorage) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	orgNamspace, err := labels.NewRequirement(typeKey, selection.Equals, []string{"organization"})
	if err != nil {
		return nil, err
	}
	options.LabelSelector = options.LabelSelector.Add(*orgNamspace)
	namespaces, err := s.namepaces.listNamespaces(ctx, options)
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

var _ rest.Updater = &organizationStorage{}
var _ rest.CreaterUpdater = &organizationStorage{}

func (s *organizationStorage) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo,
	createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {

	newOrg := &orgv1.Organization{}

	a, err := filters.GetAuthorizerAttributes(ctx)
	if err != nil {
		return nil, false, err
	}

	decision, _, err := s.namespaceAuthorizer.Authorize(ctx, a)
	if err != nil {
		return nil, false, err
	} else if decision != authorizer.DecisionAllow {
		return nil, false, kerrors.NewNotFound(newOrg.GetGroupVersionResource().GroupResource(), name)
	}

	oldOrg, err := s.Get(ctx, name, nil)
	if err != nil {

		return nil, false, fmt.Errorf("unable to get organization: %w", err)
	}

	newObj, err := objInfo.UpdatedObject(ctx, oldOrg)
	if err != nil {
		return nil, false, err
	}

	newOrg, ok := newObj.(*orgv1.Organization)
	if !ok {
		return nil, false, fmt.Errorf("new object is not an organization")
	}

	if updateValidation != nil {
		err = updateValidation(ctx, newOrg, oldOrg)
		if err != nil {
			return nil, false, err
		}
	}

	return newOrg, false, s.namepaces.updateNamespace(ctx, orgToNamespace(newOrg), options)
}

var _ rest.GracefulDeleter = &organizationStorage{}

func (s *organizationStorage) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	org, err := s.Get(ctx, name, nil)
	if err != nil {
		return nil, false, err
	}

	if deleteValidation != nil {
		err := deleteValidation(ctx, org)
		if err != nil {
			return nil, false, err
		}
	}

	ns, err := s.namepaces.deleteNamespace(ctx, orgNameToNamespaceName(name), options)
	return namespaceToOrg(ns), false, err
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
