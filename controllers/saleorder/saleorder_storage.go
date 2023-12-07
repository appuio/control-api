package saleorder

import (
	"fmt"
	"strconv"
	"strings"

	organizationv1 "github.com/appuio/control-api/apis/organization/v1"
	odooclient "github.com/appuio/go-odoo"
)

type Odoo16Credentials = odooclient.ClientConfig

type Odoo16Options struct {
	SaleOrderClientReferencePrefix string
	SaleOrderInternalNote          string
}

const defaultSaleOrderState = "sale"

type SaleOrderStorage interface {
	CreateSaleOrder(organizationv1.Organization) (string, error)
	GetSaleOrderName(organizationv1.Organization) (string, error)
}

type Odoo16Client interface {
	Read(string, []int64, *odooclient.Options, interface{}) error
	CreateSaleOrder(*odooclient.SaleOrder) (int64, error)
}

type Odoo16SaleOrderStorage struct {
	client  Odoo16Client
	options *Odoo16Options
}

func NewOdoo16Storage(credentials *Odoo16Credentials, options *Odoo16Options) (SaleOrderStorage, error) {
	client, err := odooclient.NewClient(credentials)
	return &Odoo16SaleOrderStorage{
		client:  client,
		options: options,
	}, err
}

func NewOdoo16StorageFromClient(client Odoo16Client, options *Odoo16Options) SaleOrderStorage {
	return &Odoo16SaleOrderStorage{
		client:  client,
		options: options,
	}
}

func (s *Odoo16SaleOrderStorage) CreateSaleOrder(org organizationv1.Organization) (string, error) {
	beID, err := k8sIDToOdooID(org.Spec.BillingEntityRef)
	if err != nil {
		return "", err
	}

	fetchPartnerFieldOpts := odooclient.NewOptions().FetchFields(
		"id",
		"parent_id",
	)

	beRecords := []odooclient.ResPartner{}
	err = s.client.Read(odooclient.ResPartnerModel, []int64{int64(beID)}, fetchPartnerFieldOpts, &beRecords)
	if err != nil {
		return "", fmt.Errorf("fetching accounting contact by ID: %w", err)
	}

	if len(beRecords) <= 0 {
		return "", fmt.Errorf("no results when fetching accounting contact by ID")
	}
	beRecord := beRecords[0]

	if beRecord.ParentId == nil {
		return "", fmt.Errorf("accounting contact %d has no parent", beRecord.Id.Get())
	}

	var clientRef string
	if org.Spec.DisplayName != "" {
		clientRef = fmt.Sprintf("%s (%s)", s.options.SaleOrderClientReferencePrefix, org.Spec.DisplayName)
	} else {
		clientRef = fmt.Sprintf("%s (%s)", s.options.SaleOrderClientReferencePrefix, org.ObjectMeta.Name)
	}

	newSaleOrder := odooclient.SaleOrder{
		PartnerInvoiceId: odooclient.NewMany2One(beRecord.Id.Get(), ""),
		PartnerId:        odooclient.NewMany2One(beRecord.ParentId.ID, ""),
		State:            odooclient.NewSelection(defaultSaleOrderState),
		ClientOrderRef:   odooclient.NewString(clientRef),
		InternalNote:     odooclient.NewString(s.options.SaleOrderInternalNote),
	}

	soID, err := s.client.CreateSaleOrder(&newSaleOrder)
	if err != nil {
		return "", fmt.Errorf("creating new sale order: %w", err)
	}

	return fmt.Sprint(soID), nil
}

func (s *Odoo16SaleOrderStorage) GetSaleOrderName(org organizationv1.Organization) (string, error) {
	fetchOrderFieldOpts := odooclient.NewOptions().FetchFields(
		"id",
		"name",
	)
	id, err := strconv.Atoi(org.Status.SaleOrderID)
	if err != nil {
		return "", fmt.Errorf("error parsing saleOrderID %q from organization status: %w", org.Status.SaleOrderID, err)
	}
	soRecords := []odooclient.SaleOrder{}
	err = s.client.Read(odooclient.SaleOrderModel, []int64{int64(id)}, fetchOrderFieldOpts, &soRecords)
	if err != nil {
		return "", fmt.Errorf("fetching sale order by ID: %w", err)
	}

	if len(soRecords) <= 0 {
		return "", fmt.Errorf("no results when fetching sale orders with ID %q", id)
	}

	return soRecords[0].Name.Get(), nil

}

func k8sIDToOdooID(id string) (int, error) {
	if !strings.HasPrefix(id, "be-") {
		return 0, fmt.Errorf("invalid ID, missing prefix: %s", id)
	}

	return strconv.Atoi(id[3:])
}
