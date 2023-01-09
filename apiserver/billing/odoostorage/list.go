package odoostorage

import (
	"context"

	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	"k8s.io/apimachinery/pkg/runtime"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
)

func (s billingEntityStorage) NewList() runtime.Object {
	return &billingv1.BillingEntityList{}
}

func (s *billingEntityStorage) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	bel, err := s.storage.List(ctx)
	return &billingv1.BillingEntityList{
		Items: bel,
	}, err
}
