package billingentity

import (
	"context"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
)

var _ rest.Lister = &billingEntityStorage{}

func (s billingEntityStorage) NewList() runtime.Object {
	return &billingv1.BillingEntityList{}
}

func (s *billingEntityStorage) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	return &demoentities, nil
}

var tableConvertor = rest.NewDefaultTableConvertor((&billingv1.BillingEntity{}).GetGroupVersionResource().GroupResource())

func (s *billingEntityStorage) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*v1.Table, error) {
	return tableConvertor.ConvertToTable(ctx, object, tableOptions)
}
