package odoostorage

import (
	"context"
	"errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo"
)

var _ rest.Getter = &billingEntityStorage{}

func (s *billingEntityStorage) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	be, err := s.storage.Get(ctx, name)
	if err != nil {
		if errors.Is(err, odoo.ErrNotFound) {
			return nil, apierrors.NewNotFound((&billingv1.BillingEntity{}).GetGroupVersionResource().GroupResource(), name)
		}
		return nil, err
	}

	return be, nil
}
