package odoo8

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client/model"
)

const VSHNAccountingContactNameKey = "billing.appuio.io/vshn-accounting-contact-name"

// Used to identify the accounting contact of a company.
const roleAccountCategory = 7
const companyCategory = 1

// Used to generate the UUID for the .metadata.uid field.
var metaUIDNamespace = uuid.MustParse("51887759-C769-4829-9910-BB9D5F92767D")

var roleAccountFilter = []any{"category_id", "in", []int{roleAccountCategory}}
var activeFilter = []any{"active", "in", []bool{true}}
var notInflightFilter = []any{"x_control_api_inflight", "in", []any{false}}
var mustInflightFilter = []any{"x_control_api_inflight", "not in", []any{false}}

var (
	// There's a ton of fields we don't want to override in Odoo.
	// Sadly Odoo overrides them with an empty value if the key in JSON is present even if the value is null or false.
	// The only chance to not override them is removing the key from the serialized object.
	// json:"blub,omitempty" won't omit the keys since we use custom marshalling to work around some other Odoo quirks.
	companyUpdateAllowedFields = newSet(
		"name",

		"street",
		"street2",
		"city",
		"zip",
		"country_id",

		"email",
		"phone",
	)
	accountingContactUpdateAllowedFields = newSet(
		"x_invoice_contact",
		"x_control_api_meta_status",
		"email",
	)
)

type Config struct {
	CountryIDs                   map[string]int
	AccountingContactDisplayName string
	LanguagePreference           string
	PaymentTermID                int
}

var _ odoo.OdooStorage = &Odoo8Storage{}

func NewOdoo8Storage(odooURL string, debugTransport bool, conf Config) *Odoo8Storage {
	return &Odoo8Storage{
		config: conf,
		sessionCreator: func(ctx context.Context) (client.QueryExecutor, error) {
			return client.Open(ctx, odooURL, client.ClientOptions{UseDebugLogger: debugTransport})
		},
	}
}

func NewFailedRecordScrubber(odooURL string, debugTransport bool) *FailedRecordScrubber {
	return &FailedRecordScrubber{
		sessionCreator: func(ctx context.Context) (client.QueryExecutor, error) {
			return client.Open(ctx, odooURL, client.ClientOptions{UseDebugLogger: debugTransport})
		},
	}
}

type Odoo8Storage struct {
	config Config

	sessionCreator func(ctx context.Context) (client.QueryExecutor, error)
}

type FailedRecordScrubber struct {
	sessionCreator func(ctx context.Context) (client.QueryExecutor, error)
}

func (s *Odoo8Storage) Get(ctx context.Context, name string) (*billingv1.BillingEntity, error) {
	company, accountingContact, err := s.get(ctx, name)
	if err != nil {
		return nil, err
	}

	be := mapPartnersToBillingEntity(ctx, company, accountingContact)
	return &be, nil
}

func (s *Odoo8Storage) get(ctx context.Context, name string) (company model.Partner, accountingContact model.Partner, err error) {
	id, err := k8sIDToOdooID(name)
	if err != nil {
		return model.Partner{}, model.Partner{}, err
	}

	session, err := s.sessionCreator(ctx)
	if err != nil {
		return model.Partner{}, model.Partner{}, err
	}
	o := model.NewOdoo(session)

	accountingContact, err = o.FetchPartnerByID(ctx, id, roleAccountFilter, activeFilter, notInflightFilter)
	if err != nil {
		return model.Partner{}, model.Partner{}, fmt.Errorf("fetching accounting contact by ID: %w", err)
	}

	if !accountingContact.Parent.Valid {
		return model.Partner{}, model.Partner{}, fmt.Errorf("accounting contact %d has no parent", id)
	}

	company, err = o.FetchPartnerByID(ctx, accountingContact.Parent.ID, activeFilter)
	if err != nil {
		return model.Partner{}, model.Partner{}, fmt.Errorf("fetching parent %d of accounting contact %d failed: %w", accountingContact.Parent.ID, id, err)
	}

	return company, accountingContact, nil
}

func (s *Odoo8Storage) List(ctx context.Context) ([]billingv1.BillingEntity, error) {
	l := klog.FromContext(ctx)

	session, err := s.sessionCreator(ctx)
	if err != nil {
		return nil, err
	}
	o := model.NewOdoo(session)

	accPartners, err := o.SearchPartners(ctx, []client.Filter{
		roleAccountFilter,
		activeFilter,
		notInflightFilter,
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
		bes = append(bes, mapPartnersToBillingEntity(ctx, mp, p))
	}

	return bes, nil
}

func (s *Odoo8Storage) Create(ctx context.Context, be *billingv1.BillingEntity) error {
	l := klog.FromContext(ctx)

	if be == nil {
		return errors.New("billing entity is nil")
	}
	company, accounting, err := mapBillingEntityToPartners(*be, s.config.CountryIDs)
	if err != nil {
		return fmt.Errorf("failed mapping billing entity to partners: %w", err)
	}

	inflight := uuid.New().String()
	l = l.WithValues("debug_inflight", inflight)
	company.Inflight = model.NewNullable(inflight)
	accounting.Inflight = model.NewNullable(inflight)
	setStaticCompanyFields(s.config, &company)
	setStaticAccountingContactFields(s.config, &accounting)

	session, err := s.sessionCreator(ctx)
	if err != nil {
		return err
	}
	o := model.NewOdoo(session)

	companyID, err := o.CreatePartner(ctx, company)
	if err != nil {
		return fmt.Errorf("error creating company: %w", err)
	}
	l.Info("created company (parent)", "id", companyID)

	accounting.Parent = model.OdooCompositeID{ID: companyID, Valid: true}
	accountingID, err := o.CreatePartner(ctx, accounting)
	if err != nil {
		return fmt.Errorf("error creating accounting contact: %w", err)
	}
	l.Info("created accounting contact", "id", accountingID, "parent_id", companyID)

	// reset inflight flag
	if err := o.UpdateRawPartner(ctx, []int{companyID, accountingID}, map[string]any{
		"x_control_api_inflight": false,
	}); err != nil {
		return fmt.Errorf("error resetting inflight flag: %w", err)
	}

	nbe, err := s.Get(ctx, odooIDToK8sID(accountingID))
	if err != nil {
		return fmt.Errorf("error fetching newly created billing entity: %w", err)
	}
	*be = *nbe
	return nil
}

func (s *Odoo8Storage) Update(ctx context.Context, be *billingv1.BillingEntity) error {
	l := klog.FromContext(ctx)

	if be == nil {
		return errors.New("billing entity is nil")
	}

	company, accounting, err := mapBillingEntityToPartners(*be, s.config.CountryIDs)
	if err != nil {
		return fmt.Errorf("failed mapping billing entity to partners: %w", err)
	}

	origCompany, origAccounting, err := s.get(ctx, be.Name)
	if err != nil {
		return fmt.Errorf("error fetching billing entity to update: %w", err)
	}

	session, err := s.sessionCreator(ctx)
	if err != nil {
		return err
	}
	o := model.NewOdoo(session)

	fco, err := filterFields(company, companyUpdateAllowedFields)
	if err != nil {
		return fmt.Errorf("error filtering fields: %w", err)
	}
	if err := o.UpdateRawPartner(ctx, []int{origCompany.ID}, fco); err != nil {
		return fmt.Errorf("error updating company: %w", err)
	}
	l.Info("updated company (parent)", "id", origCompany.ID)

	fac, err := filterFields(accounting, accountingContactUpdateAllowedFields)
	if err != nil {
		return fmt.Errorf("error filtering fields: %w", err)
	}
	if err := o.UpdateRawPartner(ctx, []int{origAccounting.ID}, fac); err != nil {
		return fmt.Errorf("error updating accounting contact: %w", err)
	}
	l.Info("updated accounting contact", "id", origAccounting.ID, "parent_id", origCompany.ID)

	ube, err := s.Get(ctx, odooIDToK8sID(origAccounting.ID))
	if err != nil {
		return fmt.Errorf("error fetching updated billing entity: %w", err)
	}
	*be = *ube
	return nil
}

// CleanupIncompleteRecords looks for partner records in Odoo that still have the "inflight" flag set despite being older than `minAge`. Those records are then deleted.
// Such records might come into existence due to a partially failed creation request.
func (s *FailedRecordScrubber) CleanupIncompleteRecords(ctx context.Context, minAge time.Duration) error {
	l := klog.FromContext(ctx)
	l.Info("Looking for stale inflight partner records...")

	session, err := s.sessionCreator(ctx)
	if err != nil {
		return err
	}
	o := model.NewOdoo(session)

	inflightRecords, err := o.SearchPartners(ctx, []client.Filter{
		mustInflightFilter,
	})
	if err != nil {
		return err
	}

	ids := []int{}

	for _, record := range inflightRecords {
		createdTime := record.CreationTimestamp.ToTime()

		if createdTime.Before(time.Now().Add(-1 * minAge)) {
			ids = append(ids, record.ID)
			l.Info("Preparing to delete inflight partner record", "name", record.Name, "id", record.ID)
		}
	}

	if len(ids) != 0 {
		return o.DeletePartner(ctx, ids)
	}
	return nil
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

func mapPartnersToBillingEntity(ctx context.Context, company model.Partner, accounting model.Partner) billingv1.BillingEntity {
	l := klog.FromContext(ctx)
	name := odooIDToK8sID(accounting.ID)

	var status billingv1.BillingEntityStatus
	if accounting.Status.Value != "" {
		err := json.Unmarshal([]byte(accounting.Status.Value), &status)

		if err != nil {
			l.Error(err, "Could not unmarshal BillingEntityStatus", "billingEntityName", name, "rawStatus", accounting.Status.Value)
		}
	}

	return billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			CreationTimestamp: metav1.Time{
				Time: accounting.CreationTimestamp.ToTime(),
			},
			Annotations: map[string]string{
				VSHNAccountingContactNameKey: accounting.Name,
			},
			// Since Odoo does not reuse IDs AFAIK, we can use the id from Odoo as UID.
			// Without UID patch operations will fail.
			UID: types.UID(uuid.NewSHA1(metaUIDNamespace, []byte(name)).String()),
		},
		Spec: billingv1.BillingEntitySpec{
			Name:   company.Name,
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
		Status: status,
	}
}

func mapBillingEntityToPartners(be billingv1.BillingEntity, countryIDs map[string]int) (company model.Partner, accounting model.Partner, err error) {
	countryID, ok := countryIDs[be.Spec.Address.Country]
	if !ok {
		return company, accounting, fmt.Errorf("unknown country %q", be.Spec.Address.Country)
	}

	st, err := json.Marshal(be.Status)
	if err != nil {
		return company, accounting, err
	}
	statusString := string(st)

	company = model.Partner{
		Name:  be.Spec.Name,
		Phone: model.NewNullable(be.Spec.Phone),

		Street:    model.NewNullable(be.Spec.Address.Line1),
		Street2:   model.NewNullable(be.Spec.Address.Line2),
		City:      model.NewNullable(be.Spec.Address.City),
		Zip:       model.NewNullable(be.Spec.Address.PostalCode),
		CountryID: model.NewCompositeID(countryID, ""),
	}
	company.SetEmails(be.Spec.Emails)

	accounting = model.Partner{
		InvoiceContactName: model.NewNullable(be.Spec.AccountingContact.Name),
		Status:             model.NewNullable(statusString),
	}
	accounting.SetEmails(be.Spec.AccountingContact.Emails)

	return company, accounting, nil
}

func setStaticAccountingContactFields(conf Config, a *model.Partner) {
	a.CategoryID = []int{roleAccountCategory}
	a.Name = conf.AccountingContactDisplayName
	a.Lang = model.NewNullable(conf.LanguagePreference)
	a.NotifyEmail = "always"
	a.PaymentTerm = model.OdooCompositeID{Valid: true, ID: conf.PaymentTermID}
	a.UseParentAddress = true
}

func setStaticCompanyFields(conf Config, a *model.Partner) {
	a.CategoryID = []int{companyCategory}
	a.Lang = model.NewNullable(conf.LanguagePreference)
	a.NotifyEmail = "none"
	a.PaymentTerm = model.OdooCompositeID{Valid: true, ID: conf.PaymentTermID}
}

func filterFields(p model.Partner, allowed set) (map[string]any, error) {
	sb, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	var pf map[string]any
	if err := json.Unmarshal(sb, &pf); err != nil {
		return nil, err
	}

	for k := range pf {
		if !allowed.has(k) {
			delete(pf, k)
		}
	}

	return pf, nil
}

type set map[string]struct{}

func (s set) has(key string) bool {
	_, ok := s[key]
	return ok
}

func newSet(keys ...string) set {
	s := set{}
	for _, k := range keys {
		s[k] = struct{}{}
	}
	return s
}
