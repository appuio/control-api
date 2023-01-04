package odoo

import (
	"context"
	"errors"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
)

type OdooStorage interface {
	// Create creates a new object in the storage.
	Create(ctx context.Context, be *billingv1.BillingEntity) error
	// Get retrieves an object from the storage.
	Get(ctx context.Context, name string) (*billingv1.BillingEntity, error)
	// Update updates an object in the storage.
	Update(ctx context.Context, be *billingv1.BillingEntity) error
	// List retrieves a list of objects from the storage.
	List(ctx context.Context) ([]billingv1.BillingEntity, error)
}

var ErrNotFound = errors.New("not found")
