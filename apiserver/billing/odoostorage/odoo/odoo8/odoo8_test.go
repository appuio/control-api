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

	gomock.InOrder(
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{
				{
					ID:                456,
					Name:              "Accounting",
					CreationTimestamp: client.Date(tn),
					Parent:            model.OdooCompositeID{ID: 123, Valid: true},
					EmailRaw:          model.Nullable[string]{Valid: true, Value: "accounting@test.com, notifications@test.com"},
				},
			},
		}).Return(nil),
		mock.EXPECT().SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, model.PartnerList{
			Items: []model.Partner{
				{ID: 123, Name: "Test Company"},
			},
		}).Return(nil),
	)

	s, err := subject.Get(context.Background(), "be-456")
	require.NoError(t, err)
	assert.Equal(t, &billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "be-456",
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
	subject := NewOdoo8Storage("http://localhost:8069", true)
	assert.ErrorContains(t, subject.Create(context.Background(), nil), "not implemented")
}

func TestUpdate(t *testing.T) {
	subject := NewOdoo8Storage("http://localhost:8069", true)
	assert.ErrorContains(t, subject.Update(context.Background(), nil), "not implemented")
}

func createStorage(t *testing.T) (*gomock.Controller, *clientmock.MockQueryExecutor, *oodo8Storage) {
	ctrl := gomock.NewController(t)
	mock := clientmock.NewMockQueryExecutor(ctrl)

	return ctrl, mock, &oodo8Storage{
		sessionCreator: func(ctx context.Context) (client.QueryExecutor, error) {
			return mock, nil
		},
	}
}
