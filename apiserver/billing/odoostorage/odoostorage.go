package odoostorage

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/fake"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo16"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8"
)

// NewFakeStorage returns a new storage provider for BillingEntities
func NewFakeStorage(metadataSupport bool) Storage {
	return &billingEntityStorage{
		storage: fake.NewFakeOdooStorage(metadataSupport),
	}
}

// NewOdoo8Storage returns a new storage provider for BillingEntities
func NewOdoo8Storage(odooURL string, debugTransport bool, conf odoo8.Config) Storage {
	return &billingEntityStorage{
		storage: odoo8.NewOdoo8Storage(odooURL, debugTransport, conf),
	}
}

// NewOdoo16Storage returns a new storage provider for BillingEntities
func NewOdoo16Storage(credentials odoo16.OdooCredentials, config odoo16.Config) Storage {
	return &billingEntityStorage{
		storage: odoo16.NewOdoo16Storage(credentials, config),
	}
}

type billingEntityStorage struct {
	storage odoo.OdooStorage
}

// Storage defines the features of a storage provider for BillingEntities
type Storage interface {
	rest.Storage
	rest.Scoper

	rest.CreaterUpdater
	rest.Lister
	rest.Getter
}

var _ Storage = &billingEntityStorage{}

func (s billingEntityStorage) New() runtime.Object {
	return &billingv1.BillingEntity{}
}

func (s billingEntityStorage) Destroy() {}

func (s *billingEntityStorage) NamespaceScoped() bool {
	return false
}
