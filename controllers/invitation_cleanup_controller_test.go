package controllers_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	. "github.com/appuio/control-api/controllers"
)

func Test_InvitationCleanupReconciler_Reconcile_Redeemed_Success(t *testing.T) {
	ctx := context.Background()

	subject := userv1.Invitation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
	}
	subject.Status.Token = uuid.New().String()
	subject.Status.ValidUntil = metav1.NewTime(time.Now().Add(-time.Minute))
	apimeta.SetStatusCondition(&subject.Status.Conditions, metav1.Condition{
		Type:   userv1.ConditionRedeemed,
		Status: metav1.ConditionTrue,
	})

	c := prepareTest(t, &subject)

	_, err := (&InvitationCleanupReconciler{
		Client:                c,
		Scheme:                c.Scheme(),
		Recorder:              record.NewFakeRecorder(3),
		RedeemedInvitationTTL: time.Hour,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})
	require.NoError(t, err)

	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: subject.Name}, &subject))
}

func Test_InvitationCleanupReconciler_Reconcile_RedeemedAndExpired_Success(t *testing.T) {
	ctx := context.Background()

	subject := userv1.Invitation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
	}
	subject.Status.Token = uuid.New().String()
	subject.Status.ValidUntil = metav1.NewTime(time.Now().Add(time.Minute))
	apimeta.SetStatusCondition(&subject.Status.Conditions, metav1.Condition{
		Type:   userv1.ConditionRedeemed,
		Status: metav1.ConditionTrue,
	})

	c := prepareTest(t, &subject)

	_, err := (&InvitationCleanupReconciler{
		Client:                c,
		Scheme:                c.Scheme(),
		Recorder:              record.NewFakeRecorder(3),
		RedeemedInvitationTTL: -time.Hour,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})
	require.NoError(t, err)

	require.ErrorContains(t, c.Get(ctx, types.NamespacedName{Name: subject.Name}, &subject), "not found")
}

func Test_InvitationCleanupReconciler_Reconcile_Expired_Success(t *testing.T) {
	ctx := context.Background()

	subject := userv1.Invitation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
	}
	subject.Status.Token = uuid.New().String()
	subject.Status.ValidUntil = metav1.NewTime(time.Now().Add(time.Duration(-1) * time.Minute))

	c := prepareTest(t, &subject)

	_, err := (&InvitationCleanupReconciler{
		Client:                c,
		Scheme:                c.Scheme(),
		Recorder:              record.NewFakeRecorder(3),
		RedeemedInvitationTTL: time.Hour,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})
	require.NoError(t, err)

	require.ErrorContains(t, c.Get(ctx, types.NamespacedName{Name: subject.Name}, &subject), "not found")
}

func Test_InvitationCleanupReconciler_Reconcile_Valid_Success(t *testing.T) {
	ctx := context.Background()

	subject := userv1.Invitation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
	}
	subject.Status.Token = uuid.New().String()
	subject.Status.ValidUntil = metav1.NewTime(time.Now().Add(time.Minute))

	c := prepareTest(t, &subject)

	_, err := (&InvitationCleanupReconciler{
		Client:                c,
		Scheme:                c.Scheme(),
		Recorder:              record.NewFakeRecorder(3),
		RedeemedInvitationTTL: time.Hour,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})
	require.NoError(t, err)

	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: subject.Name}, &subject))
}

func Test_InvitationCleanupReconciler_Reconcile_Uninitialized_Success(t *testing.T) {
	ctx := context.Background()

	subject := userv1.Invitation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
	}

	c := prepareTest(t, &subject)

	_, err := (&InvitationCleanupReconciler{
		Client:                c,
		Scheme:                c.Scheme(),
		Recorder:              record.NewFakeRecorder(3),
		RedeemedInvitationTTL: time.Hour,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})
	require.NoError(t, err)

	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: subject.Name}, &subject))
}
