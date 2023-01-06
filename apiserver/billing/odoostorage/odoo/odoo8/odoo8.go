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

	be := mapPartnersToBillingEntity(mainContact, &accountingContact)
	return &be, nil
}

func (s *oodo8Storage) List(ctx context.Context) ([]billingv1.BillingEntity, error) {
	l := klog.FromContext(ctx)

	session, err := s.newSession(ctx)
	if err != nil {
		return nil, err
	}
	o := model.NewOdoo(session)

	const roleAccountCategory = 7

	accPartners, err := o.SearchPartners(ctx, []client.Filter{
		[]any{"category_id", "in", []int{roleAccountCategory}},
	})
	if err != nil {
		return nil, err
	}

	mainIDs := make([]int, 0, len(accPartners))
	for _, p := range accPartners {
		if !p.Parent.Valid {
			l.Info("role account has no parent", "id", p.ID)
			continue
		}
		mainIDs = append(mainIDs, p.Parent.ID)
	}

	mainPartners, err := o.SearchPartners(ctx, []client.Filter{
		[]any{"id", "in", mainIDs},
	})
	if err != nil {
		return nil, err
	}

	mainPartnerSet := make(map[int]model.Partner)
	for _, p := range mainPartners {
		mainPartnerSet[p.ID] = p
	}

	bes := make([]billingv1.BillingEntity, 0, len(accPartners))
	for _, p := range accPartners {
		if !p.Parent.Valid {
			continue
		}
		mp, ok := mainPartnerSet[p.Parent.ID]
		if !ok {
			l.Info("could not load parent partner", "parent_id", p.Parent.ID, "id", p.ID)
			continue
		}
		bes = append(bes, mapPartnersToBillingEntity(mp, &p))
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

func mapPartnersToBillingEntity(main model.Partner, accounting *model.Partner) billingv1.BillingEntity {
	acc := billingv1.BillingEntityContact{}
	if accounting != nil {
		acc = billingv1.BillingEntityContact{
			Name:   accounting.InvoiceContactName.Value,
			Emails: accounting.Emails(),
		}
	}

	return billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name: odooIDToK8sID(accounting.ID),
		},
		Spec: billingv1.BillingEntitySpec{
			Name:   main.Name,
			Phone:  main.Phone.Value,
			Emails: main.Emails(),
			Address: billingv1.BillingEntityAddress{
				Line1:      main.Street.Value,
				Line2:      main.Street2.Value,
				City:       main.City.Value,
				PostalCode: main.Zip.Value,
				Country:    main.CountryID.Name,
			},
			AccountingContact:  acc,
			LanguagePreference: "",
		},
	}
}
