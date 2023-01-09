package model

import (
	"errors"

	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client"
)

var errNotFound = errors.New("record not found")

func IsNotFound(err error) bool {
	return errors.Is(err, errNotFound)
}

// Odoo is the developer-friendly client.Client with strongly-typed models.
type Odoo struct {
	querier client.QueryExecutor
}

// NewOdoo creates a new Odoo client.
func NewOdoo(querier client.QueryExecutor) *Odoo {
	return &Odoo{
		querier: querier,
	}
}
