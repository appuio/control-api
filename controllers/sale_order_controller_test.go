package controllers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	organizationv1 "github.com/appuio/control-api/apis/organization/v1"
	. "github.com/appuio/control-api/controllers"
	"github.com/appuio/control-api/controllers/saleorder/mock_saleorder"
)

func Test_SaleOrderReconciler_Reconcile_Create_Success(t *testing.T) {
	ctx := context.Background()
	mctrl := gomock.NewController(t)
	mock := mock_saleorder.NewMockSaleOrderStorage(mctrl)

	subject := organizationv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
		Spec: organizationv1.OrganizationSpec{
			BillingEntityRef: "be-0000",
		},
	}
	c := prepareTest(t, &subject)

	gomock.InOrder(
		mock.EXPECT().CreateSaleOrder(gomock.Any()).Return("123", nil),
	)

	_, err := (&SaleOrderReconciler{
		Client:           c,
		Scheme:           c.Scheme(),
		Recorder:         record.NewFakeRecorder(3),
		SaleOrderStorage: mock,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})

	require.NoError(t, err)
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: subject.Name}, &subject))
	require.Equal(t, "123", subject.Status.SaleOrderID)
	cond := apimeta.FindStatusCondition(subject.Status.Conditions, organizationv1.ConditionSaleOrderCreated)
	require.Equal(t, metav1.ConditionTrue, cond.Status)
}

func Test_SaleOrderReconciler_Reconcile_UpdateName_Success(t *testing.T) {
	ctx := context.Background()
	mctrl := gomock.NewController(t)
	mock := mock_saleorder.NewMockSaleOrderStorage(mctrl)

	subject := organizationv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
		Spec: organizationv1.OrganizationSpec{
			BillingEntityRef: "be-0000",
		},
		Status: organizationv1.OrganizationStatus{
			SaleOrderID: "123",
		},
	}
	c := prepareTest(t, &subject)

	gomock.InOrder(
		mock.EXPECT().GetSaleOrderName(gomock.Any()).Return("SO123", nil),
	)

	_, err := (&SaleOrderReconciler{
		Client:           c,
		Scheme:           c.Scheme(),
		Recorder:         record.NewFakeRecorder(3),
		SaleOrderStorage: mock,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})

	require.NoError(t, err)
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: subject.Name}, &subject))
	require.Equal(t, "123", subject.Status.SaleOrderID)
	require.Equal(t, "SO123", subject.Status.SaleOrderName)
	cond := apimeta.FindStatusCondition(subject.Status.Conditions, organizationv1.ConditionSaleOrderNameUpdated)
	require.Equal(t, metav1.ConditionTrue, cond.Status)
}

func Test_SaleOrderReconciler_Reconcile_NoAction_Success(t *testing.T) {
	ctx := context.Background()
	mctrl := gomock.NewController(t)
	mock := mock_saleorder.NewMockSaleOrderStorage(mctrl)

	subject := organizationv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
		Spec: organizationv1.OrganizationSpec{
			BillingEntityRef: "be-0000",
		},
		Status: organizationv1.OrganizationStatus{
			SaleOrderID:   "123",
			SaleOrderName: "SO123",
		},
	}
	c := prepareTest(t, &subject)

	mock.EXPECT().CreateSaleOrder(gomock.Any()).Times(0)
	mock.EXPECT().GetSaleOrderName(gomock.Any()).Times(0)

	_, err := (&SaleOrderReconciler{
		Client:           c,
		Scheme:           c.Scheme(),
		Recorder:         record.NewFakeRecorder(3),
		SaleOrderStorage: mock,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})

	require.NoError(t, err)
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: subject.Name}, &subject))
}

func Test_SaleOrderReconciler_Reconcile_NoBillingEntity_Success(t *testing.T) {
	ctx := context.Background()
	mctrl := gomock.NewController(t)
	mock := mock_saleorder.NewMockSaleOrderStorage(mctrl)

	subject := organizationv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
	}
	c := prepareTest(t, &subject)

	mock.EXPECT().CreateSaleOrder(gomock.Any()).Times(0)
	mock.EXPECT().GetSaleOrderName(gomock.Any()).Times(0)

	_, err := (&SaleOrderReconciler{
		Client:           c,
		Scheme:           c.Scheme(),
		Recorder:         record.NewFakeRecorder(3),
		SaleOrderStorage: mock,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})

	require.NoError(t, err)
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: subject.Name}, &subject))
}

func Test_SaleOrderReconciler_Create_Error(t *testing.T) {
	ctx := context.Background()
	mctrl := gomock.NewController(t)
	mock := mock_saleorder.NewMockSaleOrderStorage(mctrl)

	subject := organizationv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
		Spec: organizationv1.OrganizationSpec{
			BillingEntityRef: "be-0000",
		},
	}
	c := prepareTest(t, &subject)

	gomock.InOrder(
		mock.EXPECT().CreateSaleOrder(gomock.Any()).Return("", errors.New("An unanticipated fault has come to pass.")),
	)

	_, err := (&SaleOrderReconciler{
		Client:           c,
		Scheme:           c.Scheme(),
		Recorder:         record.NewFakeRecorder(3),
		SaleOrderStorage: mock,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})

	require.Error(t, err)
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: subject.Name}, &subject))
	cond := apimeta.FindStatusCondition(subject.Status.Conditions, organizationv1.ConditionSaleOrderCreated)
	require.Equal(t, metav1.ConditionFalse, cond.Status)
	require.Equal(t, organizationv1.ConditionReasonCreateFailed, cond.Reason)
	require.Equal(t, "An unanticipated fault has come to pass.", cond.Message)
}

func Test_SaleOrderReconciler_Create_StatusConditionCleared(t *testing.T) {
	ctx := context.Background()
	mctrl := gomock.NewController(t)
	mock := mock_saleorder.NewMockSaleOrderStorage(mctrl)

	subject := organizationv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
		Spec: organizationv1.OrganizationSpec{
			BillingEntityRef: "be-0000",
		},
	}
	c := prepareTest(t, &subject)

	gomock.InOrder(
		mock.EXPECT().CreateSaleOrder(gomock.Any()).Return("123", errors.New("An unanticipated fault has come to pass.")),
		mock.EXPECT().CreateSaleOrder(gomock.Any()).Return("456", nil),
		mock.EXPECT().GetSaleOrderName(gomock.Any()).Return("", errors.New("An unanticipated fault has come to pass.")),
		mock.EXPECT().GetSaleOrderName(gomock.Any()).Return("ST456", nil),
	)

	r := (&SaleOrderReconciler{
		Client:           c,
		Scheme:           c.Scheme(),
		Recorder:         record.NewFakeRecorder(3),
		SaleOrderStorage: mock,
	})

	_, err := r.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})

	require.Error(t, err)

	_, err = r.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})

	require.NoError(t, err)

	_, err = r.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})

	require.Error(t, err)

	_, err = r.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})

	require.NoError(t, err)
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: subject.Name}, &subject))
	cond := apimeta.FindStatusCondition(subject.Status.Conditions, organizationv1.ConditionSaleOrderCreated)
	require.Equal(t, metav1.ConditionTrue, cond.Status)
	require.Equal(t, "", cond.Reason)
	cond = apimeta.FindStatusCondition(subject.Status.Conditions, organizationv1.ConditionSaleOrderNameUpdated)
	require.Equal(t, metav1.ConditionTrue, cond.Status)
	require.Equal(t, "", cond.Reason)
	require.Equal(t, "456", subject.Status.SaleOrderID)
	require.Equal(t, "ST456", subject.Status.SaleOrderName)
}

func Test_SaleOrderReconciler_UpdateName_Error(t *testing.T) {
	ctx := context.Background()
	mctrl := gomock.NewController(t)
	mock := mock_saleorder.NewMockSaleOrderStorage(mctrl)

	subject := organizationv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
		Spec: organizationv1.OrganizationSpec{
			BillingEntityRef: "be-0000",
		},
		Status: organizationv1.OrganizationStatus{
			SaleOrderID: "123",
		},
	}
	c := prepareTest(t, &subject)

	gomock.InOrder(
		mock.EXPECT().GetSaleOrderName(gomock.Any()).Return("", errors.New("An unanticipated fault has come to pass.")),
	)

	_, err := (&SaleOrderReconciler{
		Client:           c,
		Scheme:           c.Scheme(),
		Recorder:         record.NewFakeRecorder(3),
		SaleOrderStorage: mock,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})

	require.Error(t, err)
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: subject.Name}, &subject))
	cond := apimeta.FindStatusCondition(subject.Status.Conditions, organizationv1.ConditionSaleOrderNameUpdated)
	require.Equal(t, metav1.ConditionFalse, cond.Status)
	require.Equal(t, organizationv1.ConditionReasonGetNameFailed, cond.Reason)
	require.Equal(t, "An unanticipated fault has come to pass.", cond.Message)
}
