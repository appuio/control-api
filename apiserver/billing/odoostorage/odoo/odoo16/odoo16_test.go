package odoo16

import (
	"context"
	"errors"
	"testing"
	"time"

	odooclient "github.com/appuio/go-odoo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo16/odoo16mock"
)

func TestGet(t *testing.T) {
	ctrl, mock, subject := createStorage(t)
	defer ctrl.Finish()

	tn := time.Now()
	st, _ := time.Parse(time.RFC3339, "2023-04-18T14:07:55Z")
	statusTime := st.Local()

	gomock.InOrder(
		mock.EXPECT().FindResPartners(gomock.Any(), gomock.Any()).Return(&odooclient.ResPartners{{
			Id:                       odooclient.NewInt(456),
			CreateDate:               odooclient.NewTime(tn),
			ParentId:                 odooclient.NewMany2One(123, ""),
			Email:                    odooclient.NewString("accounting@test.com, notifications@test.com"),
			VshnControlApiMetaStatus: odooclient.NewString("{\"conditions\":[{\"type\":\"ConditionFoo\",\"status\":\"False\",\"lastTransitionTime\":\"" + statusTime.Format(time.RFC3339) + "\",\"reason\":\"Whatever\",\"message\":\"Hello World\"}]}"),
		}}, nil),
		mock.EXPECT().FindResPartners(gomock.Any(), gomock.Any()).Return(&odooclient.ResPartners{{
			Id:   odooclient.NewInt(123),
			Name: odooclient.NewString("Test Company"),
		}}, nil),
	)

	s, err := subject.Get(context.Background(), "be-456")
	require.NoError(t, err)
	assert.Equal(t, &billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "be-456",
			UID:               "8804e682-706b-5f22-83bc-3564dadd08e1",
			CreationTimestamp: metav1.Time{Time: tn},
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
		mock.EXPECT().FindResPartners(gomock.Any(), gomock.Any()).Return(&odooclient.ResPartners{{
			Id:   odooclient.NewInt(456),
			Name: odooclient.NewString("Accounting"),
		}}, nil),
	)

	_, err := subject.Get(context.Background(), "be-456")
	require.Error(t, err)
}

func TestGet_ParentCantBeRetrieved(t *testing.T) {
	ctrl, mock, subject := createStorage(t)
	defer ctrl.Finish()

	gomock.InOrder(
		mock.EXPECT().FindResPartners(gomock.Any(), gomock.Any()).Return(&odooclient.ResPartners{{
			Id:       odooclient.NewInt(456),
			Name:     odooclient.NewString("Accounting"),
			ParentId: odooclient.NewMany2One(123, ""),
		}}, nil),
		mock.EXPECT().FindResPartners(gomock.Any(), gomock.Any()).Return(nil, errors.New("No such record")),
	)

	_, err := subject.Get(context.Background(), "be-456")
	require.ErrorContains(t, err, "fetching parent 123 of accounting contact 456 failed")
}

func TestList(t *testing.T) {
	ctrl, mock, subject := createStorage(t)
	defer ctrl.Finish()

	gomock.InOrder(
		mock.EXPECT().FindResPartners(gomock.Any(), gomock.Any()).Return(&odooclient.ResPartners{
			{
				Id:       odooclient.NewInt(456),
				ParentId: odooclient.NewMany2One(123, ""),
			},
			{
				Id:       odooclient.NewInt(457),
				ParentId: odooclient.NewMany2One(124, ""),
			},
			{
				// Can't load parent
				Id:       odooclient.NewInt(458),
				ParentId: odooclient.NewMany2One(99999, ""),
			},
			{
				// No parent
				Id:       odooclient.NewInt(459),
				ParentId: nil,
			},
		}, nil),
		mock.EXPECT().FindResPartners(gomock.Any(), gomock.Any()).Return(&odooclient.ResPartners{
			{Id: odooclient.NewInt(123), Name: odooclient.NewString("Test Company")},
			{Id: odooclient.NewInt(124), Name: odooclient.NewString("Foo Company")},
		}, nil),
	)

	s, err := subject.List(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []billingv1.BillingEntity{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "be-456",
				UID:  "8804e682-706b-5f22-83bc-3564dadd08e1",
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
				UID:  "cdb1442c-2444-5cde-8f07-7ebfa7e8825f",
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
		mock.EXPECT().CreateResPartner(gomock.Any()).Return(int64(700), nil),
		// Create accounting contact
		mock.EXPECT().CreateResPartner(gomock.Any()).Return(int64(702), nil),
		// Reset inflight flag
		mock.EXPECT().Update(odooclient.ResPartnerModel, gomock.InAnyOrder([]int64{700, 702}), gomock.Any()),
		// Fetch created company
		mock.EXPECT().FindResPartners(gomock.Any(), gomock.Any()).Return(&odooclient.ResPartners{{
			Id:         odooclient.NewInt(702),
			Name:       odooclient.NewString("Max Foobar"),
			CreateDate: odooclient.NewTime(tn),
			ParentId:   odooclient.NewMany2One(700, ""),
			Email:      odooclient.NewString("accounting@test.com, notifications@test.com"),
		}}, nil),
		mock.EXPECT().FindResPartners(gomock.Any(), gomock.Any()).Return(&odooclient.ResPartners{{
			Id:   odooclient.NewInt(700),
			Name: odooclient.NewString("Test Company"),
		}}, nil),
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
			UID:               "5ff3b076-7648-51bf-b46d-ed96cfc6f43b",
			CreationTimestamp: metav1.Time{Time: tn},
		},
		Spec: billingv1.BillingEntitySpec{
			Name:   "Test Company",
			Emails: []string{},
			AccountingContact: billingv1.BillingEntityContact{
				Name: "Max Foobar",
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
		mock.EXPECT().FindResPartners(gomock.Any(), gomock.Any()).Return(&odooclient.ResPartners{{
			Id:       odooclient.NewInt(702),
			ParentId: odooclient.NewMany2One(700, ""),
		}}, nil),
		mock.EXPECT().FindResPartners(gomock.Any(), gomock.Any()).Return(&odooclient.ResPartners{{
			Id:   odooclient.NewInt(700),
			Name: odooclient.NewString("Test Company"),
		}}, nil),
		// Update company
		mock.EXPECT().UpdateResPartner(gomock.Any()),
		// Update accounting contact
		mock.EXPECT().UpdateResPartner(gomock.Any()),
		// Fetch created company
		mock.EXPECT().FindResPartners(gomock.Any(), gomock.Any()).Return(&odooclient.ResPartners{{
			Id:                       odooclient.NewInt(702),
			CreateDate:               odooclient.NewTime(tn),
			ParentId:                 odooclient.NewMany2One(700, ""),
			Email:                    odooclient.NewString("accounting@test.com, notifications@test.com"),
			VshnControlApiMetaStatus: odooclient.NewString("{\"conditions\":[{\"type\":\"ConditionFoo\",\"status\":\"False\",\"lastTransitionTime\":\"" + statusTime.Format(time.RFC3339) + "\",\"reason\":\"Whatever\",\"message\":\"Hello World\"}]}"),
		}}, nil),
		mock.EXPECT().FindResPartners(gomock.Any(), gomock.Any()).Return(&odooclient.ResPartners{{
			Id:   odooclient.NewInt(700),
			Name: odooclient.NewString("Test Company"),
		}}, nil),
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
			UID:               "5ff3b076-7648-51bf-b46d-ed96cfc6f43b",
			CreationTimestamp: metav1.Time{Time: tn},
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

func createStorage(t *testing.T) (*gomock.Controller, *odoo16mock.MockOdoo16Client, *Odoo16Storage) {
	ctrl := gomock.NewController(t)
	mock := odoo16mock.NewMockOdoo16Client(ctrl)

	return ctrl, mock, &Odoo16Storage{
		config: Config{
			CountryIDs: map[string]int{
				"":            0,
				"Switzerland": 1,
				"Germany":     2,
			},
			LanguagePreference: "en_US",
			PaymentTermID:      2,
		},
		sessionCreator: func(ctx context.Context) (Odoo16Client, error) {
			return mock, nil
		},
	}
}

func createFailedRecordScrubber(t *testing.T) (*gomock.Controller, *odoo16mock.MockOdoo16Client, *FailedRecordScrubber) {
	ctrl := gomock.NewController(t)
	mock := odoo16mock.NewMockOdoo16Client(ctrl)

	return ctrl, mock, &FailedRecordScrubber{
		sessionCreator: func(ctx context.Context) (Odoo16Client, error) {
			return mock, nil
		},
	}
}

func TestCleanup(t *testing.T) {
	ctrl, mock, subject := createFailedRecordScrubber(t)
	defer ctrl.Finish()

	tn := time.Now()
	to := tn.Add(time.Hour * -1)

	gomock.InOrder(
		// Fetch stale records
		mock.EXPECT().FindResPartners(gomock.Any(), gomock.Any()).Return(&odooclient.ResPartners{
			{
				Id:                     odooclient.NewInt(702),
				Name:                   odooclient.NewString("Accounting"),
				CreateDate:             odooclient.NewTime(tn),
				ParentId:               odooclient.NewMany2One(700, ""),
				Email:                  odooclient.NewString("accounting@test.com, notifications@test.com"),
				VshnControlApiInflight: odooclient.NewString("fooo"),
			},
			{
				Id:                     odooclient.NewInt(703),
				Name:                   odooclient.NewString("Accounting"),
				CreateDate:             odooclient.NewTime(to),
				ParentId:               odooclient.NewMany2One(700, ""),
				Email:                  odooclient.NewString("accounting@test.com, notifications@test.com"),
				VshnControlApiInflight: odooclient.NewString("fooo"),
			},
		}, nil),
		mock.EXPECT().DeleteResPartners(gomock.Eq([]int64{703})).Return(nil),
	)

	err := subject.CleanupIncompleteRecords(context.Background(), time.Minute)
	require.NoError(t, err)

}

func Test_CachingClientCreator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	calls := 0
	shouldFail := true

	subject := CachingClientCreator(func(ctx context.Context) (Odoo16Client, error) {
		calls++
		if shouldFail {
			return nil, errors.New("failed to create client")
		}
		mock := odoo16mock.NewMockOdoo16Client(ctrl)
		mock.EXPECT().FullInitialization().Times(1).Return(nil)
		return mock, nil
	})

	// Failing call should return an error
	_, err := subject(context.Background())
	require.Error(t, err)
	assert.Equal(t, 1, calls)
	_, err = subject(context.Background())
	require.Error(t, err)
	assert.Equal(t, 2, calls)

	shouldFail = false

	// First successful call should create a new client
	client, err := subject(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 3, calls)
	prevClient := client

	// Second successful call should return the same client
	client, err = subject(context.Background())
	require.NoError(t, err)
	assert.Equal(t, prevClient, client)
	assert.Equal(t, 3, calls)
}
