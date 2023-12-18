package saleorder_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	organizationv1 "github.com/appuio/control-api/apis/organization/v1"
	"github.com/appuio/control-api/controllers/saleorder"
	"github.com/appuio/control-api/controllers/saleorder/mock_saleorder"
	odooclient "github.com/appuio/go-odoo"
)

func TestCreateCompat(t *testing.T) {
	ctrl, mock, subject := createStorageCompat(t)
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
		mock.EXPECT().CreateSaleOrder(gomock.Any()).Return(int64(149), nil),
	)

	soid, err := subject.CreateSaleOrder(organizationv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myorg",
		},
		Spec: organizationv1.OrganizationSpec{
			BillingEntityRef: "be-123",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "149", soid)
}

func TestCreate(t *testing.T) {
	ctrl, mock, subject := createStorage(t)
	defer ctrl.Finish()

	tn := time.Now()
	st, _ := time.Parse(time.RFC3339, "2023-04-18T14:07:55Z")
	statusTime := st.Local()

	gomock.InOrder(
		mock.EXPECT().Read(gomock.Any(), []int64{int64(123)}, gomock.Any(), gomock.Any()).SetArg(3, []odooclient.ResPartner{{
			Id:                       odooclient.NewInt(456),
			CreateDate:               odooclient.NewTime(tn),
			ParentId:                 odooclient.NewMany2One(123, ""),
			Email:                    odooclient.NewString("accounting@test.com, notifications@test.com"),
			VshnControlApiMetaStatus: odooclient.NewString("{\"conditions\":[{\"type\":\"ConditionFoo\",\"status\":\"False\",\"lastTransitionTime\":\"" + statusTime.Format(time.RFC3339) + "\",\"reason\":\"Whatever\",\"message\":\"Hello World\"}]}"),
		}}).Return(nil),
		mock.EXPECT().CreateSaleOrder(gomock.Any()).Return(int64(149), nil),
	)

	soid, err := subject.CreateSaleOrder(organizationv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myorg",
		},
		Spec: organizationv1.OrganizationSpec{
			BillingEntityRef: "be-123",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "149", soid)
}

func TestGet(t *testing.T) {
	ctrl, mock, subject := createStorageCompat(t)
	defer ctrl.Finish()

	gomock.InOrder(
		mock.EXPECT().Read(gomock.Any(), []int64{int64(149)}, gomock.Any(), gomock.Any()).SetArg(3, []odooclient.SaleOrder{{
			Id:   odooclient.NewInt(456),
			Name: odooclient.NewString("SO149"),
		}}).Return(nil),
	)

	soid, err := subject.GetSaleOrderName(organizationv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myorg",
		},
		Spec: organizationv1.OrganizationSpec{
			BillingEntityRef: "be-123",
		},
		Status: organizationv1.OrganizationStatus{
			SaleOrderID: "149",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "SO149", soid)
}

func TestCreateAttributesCompat(t *testing.T) {
	ctrl, mock, subject := createStorageCompat(t)
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
		mock.EXPECT().CreateSaleOrder(SaleOrderMatcher{
			PartnerId:        int64(123),
			PartnerInvoiceId: int64(456),
			State:            "sale",
			ClientOrderRef:   "client-ref (myorg)",
			InternalNote:     "internal-note",
		}).Return(int64(149), nil),
	)

	soid, err := subject.CreateSaleOrder(organizationv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myorg",
		},
		Spec: organizationv1.OrganizationSpec{
			BillingEntityRef: "be-123",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "149", soid)
}

func TestCreateAttributes(t *testing.T) {
	ctrl, mock, subject := createStorage(t)
	defer ctrl.Finish()

	tn := time.Now()
	st, _ := time.Parse(time.RFC3339, "2023-04-18T14:07:55Z")
	statusTime := st.Local()

	gomock.InOrder(
		mock.EXPECT().Read(gomock.Any(), []int64{int64(123)}, gomock.Any(), gomock.Any()).SetArg(3, []odooclient.ResPartner{{
			Id:                       odooclient.NewInt(456),
			CreateDate:               odooclient.NewTime(tn),
			ParentId:                 odooclient.NewMany2One(123, ""),
			Email:                    odooclient.NewString("accounting@test.com, notifications@test.com"),
			VshnControlApiMetaStatus: odooclient.NewString("{\"conditions\":[{\"type\":\"ConditionFoo\",\"status\":\"False\",\"lastTransitionTime\":\"" + statusTime.Format(time.RFC3339) + "\",\"reason\":\"Whatever\",\"message\":\"Hello World\"}]}"),
		}}).Return(nil),
		mock.EXPECT().CreateSaleOrder(SaleOrderMatcher{
			PartnerId:        int64(123),
			PartnerInvoiceId: int64(456),
			State:            "sale",
			ClientOrderRef:   "client-ref (myorg)",
			InternalNote:     "internal-note",
		}).Return(int64(149), nil),
	)

	soid, err := subject.CreateSaleOrder(organizationv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myorg",
		},
		Spec: organizationv1.OrganizationSpec{
			BillingEntityRef: "be-123",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "149", soid)
}

type SaleOrderMatcher struct {
	PartnerId        int64
	PartnerInvoiceId int64
	State            string
	ClientOrderRef   string
	InternalNote     string
}

func (s SaleOrderMatcher) Matches(x interface{}) bool {
	so := x.(*odooclient.SaleOrder)
	return so.PartnerId.ID == s.PartnerId && so.PartnerInvoiceId.ID == s.PartnerInvoiceId && so.State.Get() == s.State && so.ClientOrderRef.Get() == s.ClientOrderRef && so.InternalNote.Get() == s.InternalNote
}
func (s SaleOrderMatcher) String() string {
	return fmt.Sprintf("{PartnerId:%d PartnerInvoiceId:%d State:%s ClientOrderRef:%s InternalNote:%s}", s.PartnerId, s.PartnerInvoiceId, s.State, s.ClientOrderRef, s.InternalNote)
}

func createStorageCompat(t *testing.T) (*gomock.Controller, *mock_saleorder.MockOdoo16Client, saleorder.SaleOrderStorage) {
	ctrl := gomock.NewController(t)
	mock := mock_saleorder.NewMockOdoo16Client(ctrl)

	return ctrl, mock, saleorder.NewOdoo16StorageFromClient(
		mock,
		&saleorder.Odoo16Options{
			SaleOrderClientReferencePrefix: "client-ref",
			SaleOrderInternalNote:          "internal-note",
			Odoo8CompatibilityMode:         true,
		},
	)
}

func createStorage(t *testing.T) (*gomock.Controller, *mock_saleorder.MockOdoo16Client, saleorder.SaleOrderStorage) {
	ctrl := gomock.NewController(t)
	mock := mock_saleorder.NewMockOdoo16Client(ctrl)

	return ctrl, mock, saleorder.NewOdoo16StorageFromClient(
		mock,
		&saleorder.Odoo16Options{
			SaleOrderClientReferencePrefix: "client-ref",
			SaleOrderInternalNote:          "internal-note",
			Odoo8CompatibilityMode:         false,
		},
	)
}
