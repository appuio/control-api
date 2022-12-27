package billingentity

import (
	"context"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
)

var _ rest.Getter = &billingEntityStorage{}

func (s *billingEntityStorage) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	err := s.authorizer.AuthorizeGet(ctx, name)
	if err != nil {
		return nil, err
	}

	return nil, apierrors.NewMethodNotSupported((&billingv1.BillingEntity{}).GetGroupVersionResource().GroupResource(), "get")
}
