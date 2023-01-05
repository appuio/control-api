package model

import (
	"context"

	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client"
)

// Partner represents a partner ("Customer") record in Odoo
type Partner struct {
	// ID is the data record identifier.
	ID int `json:"id,omitempty" yaml:"id,omitempty"`
	// Name is the display name of the partner.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// PaymentTerm holds the terms of payment for the partner.
	PaymentTerm OdooCompositeID `json:"property_payment_term,omitempty" yaml:"property_payment_term,omitempty"`
	// ParentID is set if a customer is a sub-account (payment contact, ...) of another customer (company) account.
	Parent OdooCompositeID `json:"parent_id,omitempty" yaml:"parent_id,omitempty"`
}

// PartnerList holds the search results for Partner for deserialization
type PartnerList struct {
	Items []Partner `json:"records"`
}

// FetchPartnerByID searches for the partner by ID and returns the first entry in the result.
// If no result has been found, nil is returned without error.
func (o Odoo) FetchPartnerByID(ctx context.Context, id int) (*Partner, error) {
	result, err := o.searchPartners(ctx, []client.Filter{
		[]interface{}{"id", "in", []int{id}},
	})
	if err != nil {
		return nil, err
	}
	if len(result) > 0 {
		return &result[0], nil
	}
	// not found
	return nil, nil
}

func (o Odoo) searchPartners(ctx context.Context, domainFilters []client.Filter) ([]Partner, error) {
	result := &PartnerList{}
	err := o.querier.SearchGenericModel(ctx, client.SearchReadModel{
		Model:  "res.partner",
		Domain: domainFilters,
		Fields: []string{"name", "property_payment_term", "parent_id"},
	}, result)
	return result.Items, err
}
