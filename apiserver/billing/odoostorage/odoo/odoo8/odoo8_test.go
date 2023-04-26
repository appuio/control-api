package odoo8

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client/clientmock"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client/model"
)

func TestGet(t *testing.T) {
	ctrl, mock, subject := createStorage(t)
	defer ctrl.Finish()

	tn := time.Now()
	st, _ := time.Parse(time.RFC3339, "2023-04-18T14:07:55Z")
	statusTime := st.Local()

	gomock.InOrder(
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{
				{
					ID:                456,
					Name:              "Accounting",
					CreationTimestamp: client.Date(tn),
					Parent:            model.OdooCompositeID{ID: 123, Valid: true},
					EmailRaw:          model.Nullable[string]{Valid: true, Value: "accounting@test.com, notifications@test.com"},
					Status:            model.Nullable[string]{Valid: true, Value: "{\"conditions\":[{\"type\":\"ConditionFoo\",\"status\":\"False\",\"lastTransitionTime\":\"" + statusTime.Format(time.RFC3339) + "\",\"reason\":\"Whatever\",\"message\":\"Hello World\"}]}"},
				},
			},
		}).Return(nil),
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{
				{
					ID:   123,
					Name: "Test Company",
				},
			},
		}).Return(nil),
	)

	s, err := subject.Get(context.Background(), "be-456")
	require.NoError(t, err)
	assert.Equal(t, &billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "be-456",
			UID:               "96ac3772-d380-51b0-bf65-793ffd3837a5",
			CreationTimestamp: metav1.Time{Time: tn},
			Annotations: map[string]string{
				VSHNAccountingContactNameKey: "Accounting",
			},
		},
		Spec: billingv1.BillingEntitySpec{
			Name:   "Test Company",
			Emails: []string{},
			AccountingContact: billingv1.BillingEntityContact{
				Emails: []string{
					"accounting@test.com",
					"notifications@test.com",
				},
			},
		},
		Status: billingv1.BillingEntityStatus{
			Conditions: []metav1.Condition{
				{
					Type:               "ConditionFoo",
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.NewTime(statusTime),
					Reason:             "Whatever",
					Message:            "Hello World",
				},
			},
		},
	}, s)
}

func TestInvalidID(t *testing.T) {
	ctrl, _, subject := createStorage(t)
	defer ctrl.Finish()

	_, err := subject.Get(context.Background(), "456")
	require.Error(t, err)
	_, err = subject.Get(context.Background(), "sdf=456")
	require.Error(t, err)
}

func TestGetNoParent(t *testing.T) {
	ctrl, mock, subject := createStorage(t)
	defer ctrl.Finish()

	gomock.InOrder(
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{
				{
					ID:   456,
					Name: "Accounting",
				},
			},
		}).Return(nil),
	)

	_, err := subject.Get(context.Background(), "be-456")
	require.Error(t, err)
}

func TestGet_ParentCantBeRetrieved(t *testing.T) {
	ctrl, mock, subject := createStorage(t)
	defer ctrl.Finish()

	gomock.InOrder(
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{
				{
					ID:     456,
					Name:   "Accounting",
					Parent: model.OdooCompositeID{ID: 123, Valid: true},
				},
			},
		}).Return(nil),
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{},
		}).Return(nil),
	)

	_, err := subject.Get(context.Background(), "be-456")
	require.ErrorContains(t, err, "fetching parent 123 of accounting contact 456 failed")
}

func TestList(t *testing.T) {
	ctrl, mock, subject := createStorage(t)
	defer ctrl.Finish()

	gomock.InOrder(
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{
				{
					ID:     456,
					Name:   "Accounting",
					Parent: model.OdooCompositeID{ID: 123, Valid: true},
				},
				{
					ID:     457,
					Name:   "Accounting",
					Parent: model.OdooCompositeID{ID: 124, Valid: true},
				},
				{
					// Can't load parent
					ID:     458,
					Name:   "Accounting",
					Parent: model.OdooCompositeID{ID: 99999, Valid: true},
				},
				{
					// No parent
					ID:     459,
					Name:   "Accounting",
					Parent: model.OdooCompositeID{Valid: false},
				},
			},
		}).Return(nil),
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{
				{ID: 123, Name: "Test Company"},
				{ID: 124, Name: "Foo Company"},
			},
		}).Return(nil),
	)

	s, err := subject.List(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []billingv1.BillingEntity{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "be-456",
				Annotations: map[string]string{
					VSHNAccountingContactNameKey: "Accounting",
				},
				UID: "96ac3772-d380-51b0-bf65-793ffd3837a5",
			},
			Spec: billingv1.BillingEntitySpec{
				Name:   "Test Company",
				Emails: []string{},
				AccountingContact: billingv1.BillingEntityContact{
					Emails: []string{},
				},
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name: "be-457",
				Annotations: map[string]string{
					VSHNAccountingContactNameKey: "Accounting",
				},
				UID: "84bf1157-0532-5f6b-9257-633795440cda",
			},
			Spec: billingv1.BillingEntitySpec{
				Name:   "Foo Company",
				Emails: []string{},
				AccountingContact: billingv1.BillingEntityContact{
					Emails: []string{},
				},
			},
		},
	}, s)
}

func TestCreate(t *testing.T) {
	ctrl, mock, subject := createStorage(t)
	defer ctrl.Finish()

	tn := time.Now()

	gomock.InOrder(
		// Create company (parent)
		mock.EXPECT().CreateGenericModel(gomock.Any(), model.PartnerModel, gomock.Any()).Return(700, nil),
		// Create accounting contact
		mock.EXPECT().CreateGenericModel(gomock.Any(), model.PartnerModel, gomock.Any()).Return(702, nil),
		// Reset inflight flag
		mock.EXPECT().UpdateGenericModel(gomock.Any(), model.PartnerModel, gomock.InAnyOrder([]int{700, 702}), gomock.Any()),
		// Fetch created company
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{
				{
					ID:                702,
					Name:              "Accounting",
					CreationTimestamp: client.Date(tn),
					Parent:            model.OdooCompositeID{ID: 700, Valid: true},
					EmailRaw:          model.NewNullable("accounting@test.com, notifications@test.com"),
				},
			},
		}),
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{
				{ID: 700, Name: "Test Company"},
			},
		}),
	)

	s := &billingv1.BillingEntity{
		Spec: billingv1.BillingEntitySpec{
			Name: "Test Company",
		},
	}
	err := subject.Create(context.Background(), s)
	require.NoError(t, err)
	assert.Equal(t, &billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "be-702",
			UID:               "94362980-c246-582a-a019-817206397978",
			CreationTimestamp: metav1.Time{Time: tn},
			Annotations: map[string]string{
				VSHNAccountingContactNameKey: "Accounting",
			},
		},
		Spec: billingv1.BillingEntitySpec{
			Name:   "Test Company",
			Emails: []string{},
			AccountingContact: billingv1.BillingEntityContact{
				Emails: []string{
					"accounting@test.com",
					"notifications@test.com",
				},
			},
		},
	}, s)
}

func TestUpdate(t *testing.T) {
	ctrl, mock, subject := createStorage(t)
	defer ctrl.Finish()

	tn := time.Now()
	st, _ := time.Parse(time.RFC3339, "2023-04-18T14:07:55Z")
	statusTime := st.Local()

	gomock.InOrder(
		// Fetch existing company
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{
				{ID: 702, Parent: model.OdooCompositeID{ID: 700, Valid: true}},
			},
		}),
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{
				{ID: 700, Name: "Test Company"},
			},
		}),
		// Update company
		mock.EXPECT().UpdateGenericModel(gomock.Any(), model.PartnerModel, []int{700}, gomock.Any()),
		// Update accounting contact
		mock.EXPECT().UpdateGenericModel(gomock.Any(), model.PartnerModel, []int{702}, gomock.Any()),
		// Fetch created company
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{
				{
					ID:                702,
					Name:              "Accounting",
					CreationTimestamp: client.Date(tn),
					Parent:            model.OdooCompositeID{ID: 700, Valid: true},
					EmailRaw:          model.NewNullable("accounting@test.com, notifications@test.com"),
					Status:            model.Nullable[string]{Valid: true, Value: "{\"conditions\":[{\"type\":\"ConditionFoo\",\"status\":\"False\",\"lastTransitionTime\":\"" + statusTime.Format(time.RFC3339) + "\",\"reason\":\"Whatever\",\"message\":\"Hello World\"}]}"},
				},
			},
		}),
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{
				{
					ID:   700,
					Name: "Test Company",
				},
			},
		}),
	)

	s := &billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name: "be-702",
		},
		Spec: billingv1.BillingEntitySpec{
			Name: "Test Company",
		},
		Status: billingv1.BillingEntityStatus{
			Conditions: []metav1.Condition{
				{
					Type:               "ConditionFoo",
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.NewTime(statusTime),
					Reason:             "Whatever",
					Message:            "Hello World",
				},
			},
		},
	}
	err := subject.Update(context.Background(), s)
	require.NoError(t, err)
	assert.Equal(t, &billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "be-702",
			UID:               "94362980-c246-582a-a019-817206397978",
			CreationTimestamp: metav1.Time{Time: tn},
			Annotations: map[string]string{
				VSHNAccountingContactNameKey: "Accounting",
			},
		},
		Spec: billingv1.BillingEntitySpec{
			Name:   "Test Company",
			Emails: []string{},
			AccountingContact: billingv1.BillingEntityContact{
				Emails: []string{
					"accounting@test.com",
					"notifications@test.com",
				},
			},
		},
		Status: billingv1.BillingEntityStatus{
			Conditions: []metav1.Condition{
				{
					Type:               "ConditionFoo",
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.NewTime(statusTime),
					Reason:             "Whatever",
					Message:            "Hello World",
				},
			},
		},
	}, s)
}

func Test_CreateUpdate_UnknownCountry(t *testing.T) {
	ctrl, _, subject := createStorage(t)
	defer ctrl.Finish()

	s := &billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name: "be-702",
		},
		Spec: billingv1.BillingEntitySpec{
			Address: billingv1.BillingEntityAddress{
				Country: "Vatican City",
			},
		},
	}
	require.ErrorContains(t, subject.Create(context.Background(), s), "unknown country")
	require.ErrorContains(t, subject.Update(context.Background(), s), "unknown country")
}

func createStorage(t *testing.T) (*gomock.Controller, *clientmock.MockQueryExecutor, *Odoo8Storage) {
	ctrl := gomock.NewController(t)
	mock := clientmock.NewMockQueryExecutor(ctrl)

	return ctrl, mock, &Odoo8Storage{
		config: Config{
			CountryIDs: map[string]int{
				"":            0,
				"Switzerland": 1,
				"Germany":     2,
			},
			AccountingContactDisplayName: "Accounting",
			LanguagePreference:           "en_US",
			PaymentTermID:                2,
		},
		sessionCreator: func(ctx context.Context) (client.QueryExecutor, error) {
			return mock, nil
		},
	}
}

func TestCleanup(t *testing.T) {
	ctrl, mock, subject := createStorage(t)
	defer ctrl.Finish()

	tn := time.Now()
	to := tn.Add(time.Hour * -1)

	gomock.InOrder(
		// Fetch stale records
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{
				{
					ID:                702,
					Name:              "Accounting",
					CreationTimestamp: client.Date(tn),
					Parent:            model.OdooCompositeID{ID: 700, Valid: true},
					EmailRaw:          model.NewNullable("accounting@test.com, notifications@test.com"),
					Inflight:          model.NewNullable("fooo"),
				},
				{
					ID:                703,
					Name:              "Accounting",
					CreationTimestamp: client.Date(to),
					Parent:            model.OdooCompositeID{ID: 700, Valid: true},
					EmailRaw:          model.NewNullable("accounting@test.com, notifications@test.com"),
					Inflight:          model.NewNullable("fooo"),
				},
			},
		}),
		mock.EXPECT().DeleteGenericModel(gomock.Any(), gomock.Any(), gomock.Eq([]int{703})).Return(nil),
	)

	err := subject.CleanupIncompleteRecords(context.Background(), time.Minute)
	require.NoError(t, err)

}
