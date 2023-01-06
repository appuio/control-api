package odoo8

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client/model"
)

// Used to identify the accounting contact of a company.
const roleAccountCategory = 7

var roleAccountFilter = []any{"category_id", "in", []int{roleAccountCategory}}
var activeFilter = []any{"active", "in", []bool{true}}

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

	accountingContact, err := o.FetchPartnerByID(ctx, id, roleAccountFilter, activeFilter)
	if err != nil {
		return nil, fmt.Errorf("fetching accounting contact by ID: %w", err)
	}

	if !accountingContact.Parent.Valid {
		return nil, fmt.Errorf("accounting contact %d has no parent", id)
	}

	company, err := o.FetchPartnerByID(ctx, accountingContact.Parent.ID, activeFilter)
	if err != nil {
		return nil, fmt.Errorf("fetching parent %d of accounting contact %d failed: %w", accountingContact.Parent.ID, id, err)
	}

	be := mapPartnersToBillingEntity(company, accountingContact)
	return &be, nil
}

func (s *oodo8Storage) List(ctx context.Context) ([]billingv1.BillingEntity, error) {
	l := klog.FromContext(ctx)

	session, err := s.newSession(ctx)
	if err != nil {
		return nil, err
	}
	o := model.NewOdoo(session)

	accPartners, err := o.SearchPartners(ctx, []client.Filter{
		roleAccountFilter,
		activeFilter,
	})
	if err != nil {
		return nil, err
	}

	companyIDs := make([]int, 0, len(accPartners))
	for _, p := range accPartners {
		if !p.Parent.Valid {
			l.Info("role account has no parent", "id", p.ID)
			continue
		}
		companyIDs = append(companyIDs, p.Parent.ID)
	}

	companies, err := o.SearchPartners(ctx, []client.Filter{
		activeFilter,
		[]any{"id", "in", companyIDs},
	})
	if err != nil {
		return nil, err
	}

	companySet := make(map[int]model.Partner, len(companies))
	for _, p := range companies {
		companySet[p.ID] = p
	}

	bes := make([]billingv1.BillingEntity, 0, len(accPartners))
	for _, p := range accPartners {
		if !p.Parent.Valid {
			continue
		}
		mp, ok := companySet[p.Parent.ID]
		if !ok {
			l.Info("could not load parent partner (maybe no longer active?)", "parent_id", p.Parent.ID, "id", p.ID)
			continue
		}
		bes = append(bes, mapPartnersToBillingEntity(mp, p))
	}

	return bes, nil
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

func mapPartnersToBillingEntity(company model.Partner, accounting model.Partner) billingv1.BillingEntity {
	return billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name: odooIDToK8sID(accounting.ID),
		},
		Spec: billingv1.BillingEntitySpec{
			Name:   company.Name + ", " + accounting.Name,
			Phone:  company.Phone.Value,
			Emails: company.Emails(),
			Address: billingv1.BillingEntityAddress{
				Line1:      company.Street.Value,
				Line2:      company.Street2.Value,
				City:       company.City.Value,
				PostalCode: company.Zip.Value,
				Country:    company.CountryID.Name,
			},
			AccountingContact: billingv1.BillingEntityContact{
				Name:   accounting.InvoiceContactName.Value,
				Emails: accounting.Emails(),
			},
			LanguagePreference: "",
		},
	}
}
