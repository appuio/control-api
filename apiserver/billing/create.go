package billingentity

import (
	"context"
	"fmt"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
)

var _ rest.Creater = &billingEntityStorage{}

func (s *billingEntityStorage) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	org, ok := obj.(*billingv1.BillingEntity)
	if !ok {
		return nil, fmt.Errorf("not a billingentity: %#v", obj)
	}
	err := s.authorizer.AuthorizeContext(ctx)
	if err != nil {
		return nil, err
	}

	// Validate Org
	if err := createValidation(ctx, obj); err != nil {
		return nil, err
	}

	return nil, apierrors.NewMethodNotSupported(org.GetGroupVersionResource().GroupResource(), "create")
}
