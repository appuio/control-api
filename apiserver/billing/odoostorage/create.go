package odoostorage

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
)

var _ rest.Creater = &billingEntityStorage{}

func (s *billingEntityStorage) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	be, ok := obj.(*billingv1.BillingEntity)
	if !ok {
		return nil, fmt.Errorf("not a billingentity: %#v", obj)
	}

	// Validate Org
	if err := createValidation(ctx, obj); err != nil {
		return nil, err
	}

	return be, s.storage.Create(ctx, be)
}
