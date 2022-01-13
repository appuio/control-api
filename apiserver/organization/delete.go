package organization

import (
	"context"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
)

var _ rest.GracefulDeleter = &organizationStorage{}

func (s *organizationStorage) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	err := s.authorizer.AuthorizeContext(ctx)
	if err != nil {
		return nil, false, err
	}

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