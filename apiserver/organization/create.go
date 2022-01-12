package organization

import (
	"context"
	"fmt"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
)

var _ rest.Creater = &organizationStorage{}

func (s *organizationStorage) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	org, ok := obj.(*orgv1.Organization)
	if !ok {
		return nil, fmt.Errorf("not an organization: %#v", obj)
	}
	err := s.authorizer.AuthorizeContext(ctx)
	if err != nil {
		return nil, err
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
