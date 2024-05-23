package odoo16

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

	odooclient "github.com/appuio/go-odoo"
)

const VSHNAccountingContactNameKey = "billing.appuio.io/vshn-accounting-contact-name"

// Used to identify the accounting contact of a company.
const invoiceType = "invoice"

// Used to generate the UUID for the .metadata.uid field.
var metaUIDNamespace = uuid.MustParse("7550b1ae-7a2a-485e-a75d-6f931b2cd73f")

var activeFilter = odooclient.NewCriterion("active", "=", true)
var invoiceTypeFilter = odooclient.NewCriterion("type", "=", invoiceType)
var notInflightFilter = odooclient.NewCriterion("vshn_control_api_inflight", "=", false)
var mustInflightFilter = odooclient.NewCriterion("vshn_control_api_inflight", "!=", false)

var fetchPartnerFieldOpts = odooclient.NewOptions().FetchFields(
	"id",
	"type",
	"name",
	"display_name",
	"country_id",
	"commercial_partner_id",
	"contact_address",

	"child_ids",
	"user_ids",

	"email",
	"phone",
	"street",
	"street2",
	"city",
	"zip",
	"country_id",

	"parent_id",
	"vshn_control_api_meta_status",
	"vshn_control_api_inflight",
)

type OdooCredentials = odooclient.ClientConfig

type Config struct {
	CountryIDs         map[string]int
	LanguagePreference string
	PaymentTermID      int
}

var _ odoo.OdooStorage = &Odoo16Storage{}

func NewOdoo16Storage(credentials OdooCredentials, conf Config) *Odoo16Storage {
	return &Odoo16Storage{
		config: conf,
		sessionCreator: func(ctx context.Context) (Odoo16Client, error) {
			return odooclient.NewClient(&credentials)
		},
	}
}

func NewFailedRecordScrubber(credentials OdooCredentials) *FailedRecordScrubber {
	return &FailedRecordScrubber{
		sessionCreator: func(ctx context.Context) (Odoo16Client, error) {
			return odooclient.NewClient(&credentials)
		},
	}
}

type Odoo16Storage struct {
	config Config

	sessionCreator func(ctx context.Context) (Odoo16Client, error)
}

type FailedRecordScrubber struct {
	sessionCreator func(ctx context.Context) (Odoo16Client, error)
}

//go:generate go run go.uber.org/mock/mockgen -destination=./odoo16mock/$GOFILE -package odoo16mock . Odoo16Client
type Odoo16Client interface {
	Update(string, []int64, interface{}) error
	FindResPartners(*odooclient.Criteria, *odooclient.Options) (*odooclient.ResPartners, error)
	CreateResPartner(*odooclient.ResPartner) (int64, error)
	UpdateResPartner(*odooclient.ResPartner) error
	DeleteResPartners([]int64) error
}

func (s *Odoo16Storage) Get(ctx context.Context, name string) (*billingv1.BillingEntity, error) {
	company, accountingContact, err := s.get(ctx, name)
	if err != nil {
		return nil, err
	}

	be := mapPartnersToBillingEntity(ctx, company, accountingContact)
	return &be, nil
}

func (s *Odoo16Storage) get(ctx context.Context, name string) (company odooclient.ResPartner, accountingContact odooclient.ResPartner, err error) {
	id, err := k8sIDToOdooID(name)
	if err != nil {
		return odooclient.ResPartner{}, odooclient.ResPartner{}, err
	}

	session, err := s.sessionCreator(ctx)
	if err != nil {
		return odooclient.ResPartner{}, odooclient.ResPartner{}, err
	}

	accp, err := session.FindResPartners(
		newValidInvoiceRecordCriteria().AddCriterion(odooclient.NewCriterion("id", "=", id)),
		fetchPartnerFieldOpts)
	if err != nil {
		return odooclient.ResPartner{}, odooclient.ResPartner{}, fmt.Errorf("error fetching accounting contact %d: %w", id, err)
	}
	if accp == nil {
		return odooclient.ResPartner{}, odooclient.ResPartner{}, fmt.Errorf("fetching accounting contact %d returned nil", id)
	}
	acc := *accp
	if len(acc) <= 0 {
		return odooclient.ResPartner{}, odooclient.ResPartner{}, fmt.Errorf("no results when fetching accounting contact %d", id)
	}
	if len(acc) > 1 {
		return odooclient.ResPartner{}, odooclient.ResPartner{}, fmt.Errorf("multiple results when fetching accounting contact %d", id)
	}
	accountingContact = acc[0]

	if accountingContact.ParentId == nil {
		return odooclient.ResPartner{}, odooclient.ResPartner{}, fmt.Errorf("accounting contact %d has no parent", id)
	}

	cpp, err := session.FindResPartners(
		odooclient.NewCriteria().AddCriterion(activeFilter).AddCriterion(odooclient.NewCriterion("id", "=", accountingContact.ParentId.Get())),
		fetchPartnerFieldOpts)
	if err != nil {
		return odooclient.ResPartner{}, odooclient.ResPartner{}, fmt.Errorf("fetching parent %d of accounting contact %d failed: %w", accountingContact.ParentId.ID, id, err)
	}
	if cpp == nil {
		return odooclient.ResPartner{}, odooclient.ResPartner{}, fmt.Errorf("fetching parent %d of accounting contact %d returned nil", accountingContact.ParentId.ID, id)
	}
	cp := *cpp
	if len(cp) <= 0 {
		return odooclient.ResPartner{}, odooclient.ResPartner{}, fmt.Errorf("no results when fetching parent %d of accounting contact %d", accountingContact.ParentId.ID, id)
	}
	if len(cp) > 1 {
		return odooclient.ResPartner{}, odooclient.ResPartner{}, fmt.Errorf("multiple results when fetching parent %d of accounting contact %d", accountingContact.ParentId.ID, id)
	}
	company = cp[0]

	return company, accountingContact, nil
}

func (s *Odoo16Storage) List(ctx context.Context) ([]billingv1.BillingEntity, error) {
	l := klog.FromContext(ctx)

	session, err := s.sessionCreator(ctx)
	if err != nil {
		return nil, err
	}

	accPartners, err := session.FindResPartners(newValidInvoiceRecordCriteria(), fetchPartnerFieldOpts)
	if err != nil {
		return nil, err
	}

	companyIDs := make([]int, 0, len(*accPartners))
	for _, p := range *accPartners {
		if p.ParentId == nil {
			l.Info("role account has no parent", "id", p.Id)
			continue
		}
		companyIDs = append(companyIDs, int(p.ParentId.ID))
	}

	criteria := odooclient.NewCriteria().AddCriterion(activeFilter).AddCriterion(odooclient.NewCriterion("id", "in", companyIDs))
	companies, err := session.FindResPartners(criteria, fetchPartnerFieldOpts)
	if err != nil {
		return nil, err
	}

	companySet := make(map[int]odooclient.ResPartner, len(*companies))
	for _, p := range *companies {
		companySet[int(p.Id.Get())] = p
	}

	bes := make([]billingv1.BillingEntity, 0, len(*accPartners))
	for _, p := range *accPartners {
		if p.ParentId == nil {
			continue
		}
		mp, ok := companySet[int(p.ParentId.ID)]
		if !ok {
			l.Info("could not load parent partner (maybe no longer active?)", "parent_id", p.ParentId.ID, "id", p.Id.Get())
			continue
		}
		bes = append(bes, mapPartnersToBillingEntity(ctx, mp, p))
	}

	return bes, nil
}

func (s *Odoo16Storage) Create(ctx context.Context, be *billingv1.BillingEntity) error {
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
	company.VshnControlApiInflight = odooclient.NewString(inflight)
	accounting.VshnControlApiInflight = odooclient.NewString(inflight)
	setStaticCompanyFields(s.config, &company)
	setStaticAccountingContactFields(s.config, &accounting)

	session, err := s.sessionCreator(ctx)
	if err != nil {
		return err
	}

	l.Info("about to create partner")
	companyID, err := session.CreateResPartner(&company)
	if err != nil {
		return fmt.Errorf("error creating company: %w", err)
	}
	l.Info("created company (parent)", "id", companyID)

	accounting.ParentId = odooclient.NewMany2One(companyID, "")
	accountingID, err := session.CreateResPartner(&accounting)
	if err != nil {
		return fmt.Errorf("error creating accounting contact: %w", err)
	}
	l.Info("created accounting contact", "id", accountingID, "parent_id", companyID)

	// reset inflight flag
	if err := session.Update(odooclient.ResPartnerModel, []int64{companyID, accountingID}, map[string]any{
		"vshn_control_api_inflight": false,
	}); err != nil {
		return fmt.Errorf("error resetting inflight flag: %w", err)
	}

	nbe, err := s.Get(ctx, odooIDToK8sID(int(accountingID)))
	if err != nil {
		return fmt.Errorf("error fetching newly created billing entity: %w", err)
	}
	*be = *nbe
	return nil
}

func (s *Odoo16Storage) Update(ctx context.Context, be *billingv1.BillingEntity) error {
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

	company.Id = origCompany.Id
	accounting.Id = origAccounting.Id

	if err := session.UpdateResPartner(&company); err != nil {
		return fmt.Errorf("error updating company: %w", err)
	}
	l.Info("updated company (parent)", "id", origCompany.Id.Get())

	if err := session.UpdateResPartner(&accounting); err != nil {
		return fmt.Errorf("error updating accounting contact: %w", err)
	}
	l.Info("updated accounting contact", "id", origAccounting.Id.Get(), "parent_id", origCompany.Id.Get())

	ube, err := s.Get(ctx, odooIDToK8sID(int(origAccounting.Id.Get())))
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

	inflightRecords, err := session.FindResPartners(odooclient.NewCriteria().AddCriterion(mustInflightFilter), fetchPartnerFieldOpts)
	if err != nil {
		return err
	}

	ids := []int64{}

	for _, record := range *inflightRecords {
		createdTime := record.CreateDate.Get()

		if createdTime.Before(time.Now().Add(-1 * minAge)) {
			ids = append(ids, record.Id.Get())
			l.Info("Preparing to delete inflight partner record", "name", record.Name, "id", record.Id.Get())
		}
	}

	if len(ids) != 0 {
		return session.DeleteResPartners(ids)
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

func mapPartnersToBillingEntity(ctx context.Context, company odooclient.ResPartner, accounting odooclient.ResPartner) billingv1.BillingEntity {
	l := klog.FromContext(ctx)
	name := odooIDToK8sID(int(accounting.Id.Get()))

	var status billingv1.BillingEntityStatus
	if accounting.VshnControlApiMetaStatus.Get() != "" {
		err := json.Unmarshal([]byte(accounting.VshnControlApiMetaStatus.Get()), &status)

		if err != nil {
			l.Error(err, "Could not unmarshal BillingEntityStatus", "billingEntityName", name, "rawStatus", accounting.VshnControlApiMetaStatus.Get())
		}
	}

	var country string
	if company.CountryId != nil {
		country = company.CountryId.Name
	}
	return billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			CreationTimestamp: metav1.Time{
				Time: accounting.CreateDate.Get(),
			},
			// Since Odoo does not reuse IDs AFAIK, we can use the id from Odoo as UID.
			// Without UID patch operations will fail.
			UID: types.UID(uuid.NewSHA1(metaUIDNamespace, []byte(name)).String()),
		},
		Spec: billingv1.BillingEntitySpec{
			Name:   company.Name.Get(),
			Phone:  company.Phone.Get(),
			Emails: splitCommaSeparated(company.Email.Get()),
			Address: billingv1.BillingEntityAddress{
				Line1:      company.Street.Get(),
				Line2:      company.Street2.Get(),
				City:       company.City.Get(),
				PostalCode: company.Zip.Get(),
				Country:    country,
			},
			AccountingContact: billingv1.BillingEntityContact{
				Name:   accounting.Name.Get(),
				Emails: splitCommaSeparated(accounting.Email.Get()),
			},
			LanguagePreference: "",
		},
		Status: status,
	}
}

func mapBillingEntityToPartners(be billingv1.BillingEntity, countryIDs map[string]int) (company odooclient.ResPartner, accounting odooclient.ResPartner, err error) {
	countryID, ok := countryIDs[be.Spec.Address.Country]
	if !ok {
		return company, accounting, fmt.Errorf("unknown country %q", be.Spec.Address.Country)
	}

	st, err := json.Marshal(be.Status)
	if err != nil {
		return company, accounting, err
	}
	statusString := string(st)

	company = odooclient.ResPartner{
		Name:  odooclient.NewString(be.Spec.Name),
		Phone: odooclient.NewString(be.Spec.Phone),

		Street:    odooclient.NewString(be.Spec.Address.Line1),
		Street2:   odooclient.NewString(be.Spec.Address.Line2),
		City:      odooclient.NewString(be.Spec.Address.City),
		Zip:       odooclient.NewString(be.Spec.Address.PostalCode),
		CountryId: odooclient.NewMany2One(int64(countryID), ""),
		Email:     odooclient.NewString(strings.Join(be.Spec.Emails, ", ")),
	}

	accounting = odooclient.ResPartner{
		Name:                     odooclient.NewString(be.Spec.AccountingContact.Name),
		VshnControlApiMetaStatus: odooclient.NewString(statusString),
		Email:                    odooclient.NewString(strings.Join(be.Spec.AccountingContact.Emails, ", ")),
	}

	return company, accounting, nil
}

func setStaticAccountingContactFields(conf Config, a *odooclient.ResPartner) {
	a.Lang = odooclient.NewSelection(conf.LanguagePreference)
	a.Type = odooclient.NewSelection(invoiceType)
	a.PropertyPaymentTermId = odooclient.NewMany2One(int64(conf.PaymentTermID), "")
}

func setStaticCompanyFields(conf Config, a *odooclient.ResPartner) {
	a.Lang = odooclient.NewSelection(conf.LanguagePreference)
	a.PropertyPaymentTermId = odooclient.NewMany2One(int64(conf.PaymentTermID), "")
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

func newValidInvoiceRecordCriteria() *odooclient.Criteria {
	return odooclient.NewCriteria().
		AddCriterion(invoiceTypeFilter).
		AddCriterion(activeFilter).
		AddCriterion(notInflightFilter)
}
