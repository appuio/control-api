package odoostorage

import (
	"context"

	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"

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

var tableConvertor = rest.NewDefaultTableConvertor((&billingv1.BillingEntity{}).GetGroupVersionResource().GroupResource())

func (s *billingEntityStorage) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return tableConvertor.ConvertToTable(ctx, object, tableOptions)
}
