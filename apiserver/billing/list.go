package billingentity

import (
	"context"

	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
)

var _ rest.Lister = &billingEntityStorage{}

func (s billingEntityStorage) NewList() runtime.Object {
	return &billingv1.BillingEntityList{}
}

func (s *billingEntityStorage) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	err := s.authorizer.AuthorizeContext(ctx)
	if err != nil {
		return nil, err
	}

	bel, err := s.storage.List(ctx)
	filtered := make([]billingv1.BillingEntity, 0, len(bel))
	for _, be := range bel {
		err := s.authorizer.AuthorizeGet(ctx, be.Name)
		if err != nil {
			continue
		}
		filtered = append(filtered, be)
	}

	return &billingv1.BillingEntityList{
		Items: filtered,
	}, err
}

var tableConvertor = rest.NewDefaultTableConvertor((&billingv1.BillingEntity{}).GetGroupVersionResource().GroupResource())

func (s *billingEntityStorage) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return tableConvertor.ConvertToTable(ctx, object, tableOptions)
}
