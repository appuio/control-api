package odoo8

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client/model"
)

func NewOdoo8Storage(odooURL string, debugTransport bool) odoo.OdooStorage {
	return &oodo8Storage{
		odooURL:        odooURL,
		debugTransport: debugTransport,
	}
}

type oodo8Storage struct {
	odooURL        string
	debugTransport bool
}

func (s *oodo8Storage) Get(ctx context.Context, name string) (*billingv1.BillingEntity, error) {
	id, err := k8sIDToOdooID(name)
	if err != nil {
		return nil, err
	}

	session, err := s.newSession(ctx)
	if err != nil {
		return nil, err
	}
	o := model.NewOdoo(session)

	accountingContact, err := o.FetchPartnerByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching accounting contact by ID: %w", err)
	}

	if !accountingContact.Parent.Valid {
		return nil, fmt.Errorf("accounting contact %d has no parent", id)
	}

	mainContact, err := o.FetchPartnerByID(ctx, accountingContact.Parent.ID)
	if err != nil {
		return nil, fmt.Errorf("fetching accounting contact by ID: %w", err)
	}

	return &billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: billingv1.BillingEntitySpec{
			Name:   mainContact.Name,
			Phone:  mainContact.Phone,
			Emails: mainContact.Emails(),
			Address: billingv1.BillingEntityAddress{
				Line1:      mainContact.Street,
				Line2:      mainContact.Street2.Value,
				City:       mainContact.City,
				PostalCode: mainContact.Zip,
				Country:    mainContact.CountryID.Name,
			},
			AccountingContact: billingv1.BillingEntityContact{
				Name:   accountingContact.InvoiceContactName.Value,
				Emails: accountingContact.Emails(),
			},
			LanguagePreference: "",
		},
	}, nil
}

func (s *oodo8Storage) List(ctx context.Context) ([]billingv1.BillingEntity, error) {
	return []billingv1.BillingEntity{}, nil
}

func (s *oodo8Storage) Create(ctx context.Context, be *billingv1.BillingEntity) error {
	return errors.New("not implemented")
}

func (s *oodo8Storage) Update(ctx context.Context, be *billingv1.BillingEntity) error {
	return errors.New("not implemented")
}

func (s *oodo8Storage) newSession(ctx context.Context) (*client.Session, error) {
	return client.Open(ctx, s.odooURL, client.ClientOptions{UseDebugLogger: s.debugTransport})
}

func k8sIDToOdooID(id string) (int, error) {
	if !strings.HasPrefix(id, "be-") {
		return 0, fmt.Errorf("invalid ID, missing prefix: %s", id)
	}

	return strconv.Atoi(id[3:])
}

func odooIDToK8sID(id int) string {
	return fmt.Sprintf("be-%d", id)
}
