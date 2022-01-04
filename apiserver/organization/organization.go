package organization

import (
	"context"
	"errors"
	"fmt"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/watch"
	genericregistry "k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	restbuilder "sigs.k8s.io/apiserver-runtime/pkg/builder/rest"
)

// New returns a new storage provider for Organizations
func New() restbuilder.ResourceHandlerProvider {
	return func(s *runtime.Scheme, g genericregistry.RESTOptionsGetter) (rest.Storage, error) {
		return &organizationStorage{
			namepaces: &loopbackNamespaceProvider{},
		}, nil
	}
}

type organizationStorage struct {
	namepaces namespaceProvider
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
	org := &orgv1.Organization{}
	ns, err := s.namepaces.GetNamespace(ctx, name, options)
	if err != nil {
		return nil, convertNamespaceError(err)
	}
	org = orgv1.NewOrganizationFromNS(ns)
	if org == nil {
		// This namespace is not an organization
		return nil, apierrors.NewNotFound(org.GetGroupVersionResource().GroupResource(), name)
	}
	return org, nil
}

var _ rest.Creater = &organizationStorage{}

func (s *organizationStorage) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	org, ok := obj.(*orgv1.Organization)
	if !ok {
		return nil, fmt.Errorf("not an organization: %#v", obj)
	}

	// Validate Org
	if err := createValidation(ctx, obj); err != nil {
		return nil, err
	}

	if err := s.namepaces.CreateNamespace(ctx, org.ToNamespace(), options); err != nil {
		return nil, convertNamespaceError(err)
	}
	return org, nil
}

var _ rest.Lister = &organizationStorage{}

func (s organizationStorage) NewList() runtime.Object {
	return &orgv1.OrganizationList{}
}

func (s *organizationStorage) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	orgNamspace, err := labels.NewRequirement(orgv1.TypeKey, selection.Equals, []string{orgv1.OrgType})
	if err != nil {
		return nil, err
	}
	options.LabelSelector = options.LabelSelector.Add(*orgNamspace)
	namespaces, err := s.namepaces.ListNamespaces(ctx, options)
	if err != nil {
		return nil, err
	}

	res := orgv1.OrganizationList{
		ListMeta: namespaces.ListMeta,
		Items:    []orgv1.Organization{},
	}

	for _, n := range namespaces.Items {
		res.Items = append(res.Items, *orgv1.NewOrganizationFromNS(&n))
	}

	return &res, nil
}

var _ rest.Updater = &organizationStorage{}
var _ rest.CreaterUpdater = &organizationStorage{}

func (s *organizationStorage) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo,
	createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {

	newOrg := &orgv1.Organization{}

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

	return newOrg, false, s.namepaces.UpdateNamespace(ctx, newOrg.ToNamespace(), options)
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

	ns, err := s.namepaces.DeleteNamespace(ctx, name, options)
	return orgv1.NewOrganizationFromNS(ns), false, convertNamespaceError(err)
}

var _ rest.Watcher = &organizationStorage{}

func (s *organizationStorage) Watch(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	orgNamspace, err := labels.NewRequirement(orgv1.TypeKey, selection.Equals, []string{orgv1.OrgType})
	if err != nil {
		return nil, err
	}
	options.LabelSelector = options.LabelSelector.Add(*orgNamspace)

	nsWatcher, err := s.namepaces.WatchNamespaces(ctx, options)
	if err != nil {
		return nil, err
	}

	return watch.Filter(nsWatcher, func(in watch.Event) (out watch.Event, keep bool) {
		if in.Object == nil {
			// This should never happen, let downstream deal with it
			return in, true
		}
		ns, ok := in.Object.(*corev1.Namespace)
		if !ok {
			// We received a non Namespace object
			// This is most likely an error so we pass it on
			return in, true
		}

		in.Object = orgv1.NewOrganizationFromNS(ns)

		return in, true
	}), nil
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
