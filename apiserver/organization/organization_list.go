package organization

import (
	"context"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	corev1 "k8s.io/api/core/v1"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/registry/rest"
)

var _ rest.Lister = &organizationStorage{}

func (s organizationStorage) NewList() runtime.Object {
	return &orgv1.OrganizationList{}
}

func (s *organizationStorage) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	namespaces, err := s.namepaces.ListNamespaces(ctx, extendNamespaceListOption(options))
	if err != nil {
		return nil, convertNamespaceError(err)
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

var _ rest.Watcher = &organizationStorage{}

func (s *organizationStorage) Watch(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {

	nsWatcher, err := s.namepaces.WatchNamespaces(ctx, extendNamespaceListOption(options))
	if err != nil {
		return nil, convertNamespaceError(err)
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

func extendNamespaceListOption(options *metainternalversion.ListOptions) *metainternalversion.ListOptions {
	orgNamspace, err := labels.NewRequirement(orgv1.TypeKey, selection.Equals, []string{orgv1.OrgType})
	if err != nil {
		// The input is static. This call will only fail during development.
		panic(err)
	}
	if options == nil {
		options = &metainternalversion.ListOptions{}
	}
	if options.LabelSelector == nil {
		options.LabelSelector = labels.NewSelector()
	}
	options.LabelSelector = options.LabelSelector.Add(*orgNamspace)

	return options
}
