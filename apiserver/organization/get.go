package organization

import (
	"context"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
)

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
