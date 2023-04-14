package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client"
)

// PartnerModel is the name of the Odoo model for partners.
const PartnerModel = "res.partner"

// Partner represents a partner ("Customer") record in Odoo
type Partner struct {
	// ID is the data record identifier.
	ID int `json:"id,omitempty" yaml:"id,omitempty"`
	// Name is the display name of the partner.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// CreationTimestamp is the creation date of the partner.
	CreationTimestamp client.Date `json:"create_date,omitempty" yaml:"create_date,omitempty"`

	// CategoryID is the category of the partner.
	CategoryID CategoryIDs `json:"category_id,omitempty" yaml:"category_id,omitempty"`
	// Lang is the language of the partner.
	Lang Nullable[string] `json:"lang,omitempty" yaml:"lang,omitempty"`
	// NotifyEmail is the email notification preference of the partner.
	NotifyEmail string `json:"notify_email,omitempty" yaml:"notify_email,omitempty"`
	// ParentID is set if a customer is a sub-account (payment contact, ...) of another customer (company) account.
	Parent OdooCompositeID `json:"parent_id,omitempty" yaml:"parent_id,omitempty"`
	// PaymentTerm holds the terms of payment for the partner.
	PaymentTerm OdooCompositeID `json:"property_payment_term,omitempty" yaml:"property_payment_term,omitempty"`

	// InvoiceContactName is the contact person for invoices.
	InvoiceContactName Nullable[string] `json:"x_invoice_contact,omitempty" yaml:"x_invoice_contact,omitempty"`
	// UseParentAddress is set if the partner uses the address of the parent partner.
	UseParentAddress bool `json:"use_parent_address,omitempty" yaml:"use_parent_address,omitempty"`

	// Street is the street address of the partner.
	Street Nullable[string] `json:"street,omitempty" yaml:"street,omitempty"`
	// Street2 is the second line of the street address of the partner.
	Street2 Nullable[string] `json:"street2,omitempty" yaml:"street2,omitempty"`
	// City is the city of the partner.
	City Nullable[string] `json:"city,omitempty" yaml:"city,omitempty"`
	// Zip is the zip code of the partner.
	Zip Nullable[string] `json:"zip,omitempty" yaml:"zip,omitempty"`
	// CountryID is the country of the partner.
	CountryID OdooCompositeID `json:"country_id,omitempty" yaml:"country_id,omitempty"`

	// EmailRaw is the email addresses of the partner, comma-separated.
	EmailRaw Nullable[string] `json:"email,omitempty" yaml:"email,omitempty"`
	// Phone is the phone number of the partner.
	Phone Nullable[string] `json:"phone,omitempty" yaml:"phone,omitempty"`

	// Inflight allows detecting half-finished creates.
	Inflight Nullable[string] `json:"x_control_api_inflight,omitempty" yaml:"x_control_api_inflight,omitempty"`
}

func (p Partner) Emails() []string {
	return splitCommaSeparated(p.EmailRaw.Value)
}

func (p *Partner) SetEmails(emails []string) {
	p.EmailRaw = NewNullable(strings.Join(emails, ", "))
}

func splitCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}
	p := strings.Split(s, ",")
	for i, v := range p {
		p[i] = strings.TrimSpace(v)
	}
	return p
}

// PartnerList holds the search results for Partner for deserialization
type PartnerList struct {
	Items []Partner `json:"records"`
}

// PartnerFields is the list of fields that are fetched for a partner.
var PartnerFields = []string{
	"name",
	"create_date",

	"category_id",
	"lang",
	"notify_email",
	"parent_id",
	"property_payment_term",

	"x_invoice_contact",
	"use_parent_address",

	"street",
	"street2",
	"city",
	"zip",
	"country_id",

	"email",
	"phone",
}

// FetchPartnerByID searches for the partner by ID and returns the first entry in the result.
// If no result has been found, a not found error is returned.
func (o Odoo) FetchPartnerByID(ctx context.Context, id int, domainFilters ...client.Filter) (Partner, error) {
	result, err := o.SearchPartners(ctx, append(domainFilters,
		[]any{"id", "in", []int{id}},
	))
	if err != nil {
		return Partner{}, err
	}
	if len(result) > 0 {
		return result[0], nil
	}
	return Partner{}, fmt.Errorf("partner with ID %d: %w", id, errNotFound)
}

func (o Odoo) SearchPartners(ctx context.Context, domainFilters []client.Filter) ([]Partner, error) {
	result := &PartnerList{}
	err := o.querier.SearchGenericModel(ctx, client.SearchReadModel{
		Model:  PartnerModel,
		Domain: domainFilters,
		Fields: PartnerFields,
	}, result)
	return result.Items, err
}

func (o Odoo) CreatePartner(ctx context.Context, p Partner) (id int, err error) {
	id, err = o.querier.CreateGenericModel(ctx, PartnerModel, p)
	return id, err
}

func (o Odoo) UpdateRawPartner(ctx context.Context, ids []int, raw any) error {
	return o.querier.UpdateGenericModel(ctx, PartnerModel, ids, raw)
}
