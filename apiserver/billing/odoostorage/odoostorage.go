package odoostorage

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/fake"
)

// New returns a new storage provider for Organizations
func New() rest.Storage {
	return &billingEntityStorage{
		storage: fake.NewFakeOdooStorage(false),
	}
}

type billingEntityStorage struct {
	storage odoo.OdooStorage
}

var _ rest.Storage = &billingEntityStorage{}

func (s billingEntityStorage) New() runtime.Object {
	return &billingv1.BillingEntity{}
}

func (s billingEntityStorage) Destroy() {}

var _ rest.Scoper = &billingEntityStorage{}

func (s *billingEntityStorage) NamespaceScoped() bool {
	return false
}
