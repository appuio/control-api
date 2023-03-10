package controllers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	. "github.com/appuio/control-api/controllers"
)

type FailingSender struct{}
type SenderWithConstantId struct{}

func (f *FailingSender) Send(context.Context, string, userv1.Invitation) (string, error) {
	return "", errors.New("Err0r")
}

func (s *SenderWithConstantId) Send(context.Context, string, userv1.Invitation) (string, error) {
	return "ID10", nil
}

func Test_InvitationEmailReconciler_Reconcile_Success(t *testing.T) {
	ctx := context.Background()

	subject := baseInvitation()
	apimeta.SetStatusCondition(&subject.Status.Conditions, metav1.Condition{
		Type:   userv1.ConditionEmailSent,
		Status: metav1.ConditionFalse,
	})

	c := prepareTest(t, subject)

	r := invitationEmailReconciler(c)
	_, err := r.Reconcile(ctx, requestFor(subject))
	require.NoError(t, err)

	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(subject), subject))
	require.True(t, apimeta.IsStatusConditionTrue(subject.Status.Conditions, userv1.ConditionEmailSent))
	condition := apimeta.FindStatusCondition(subject.Status.Conditions, userv1.ConditionEmailSent)
	require.Equal(t, "Message ID: ID10", condition.Message)
}

func Test_InvitationEmailReconciler_Reconcile_WithSendingFailure_Success(t *testing.T) {
	ctx := context.Background()

	subject := baseInvitation()
	apimeta.SetStatusCondition(&subject.Status.Conditions, metav1.Condition{
		Type:   userv1.ConditionEmailSent,
		Status: metav1.ConditionFalse,
	})

	c := prepareTest(t, subject)

	r := invitationEmailReconcilerWithFailingSender(c)
	_, err := r.Reconcile(ctx, requestFor(subject))
	require.Error(t, err)

	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(subject), subject))
	require.False(t, apimeta.IsStatusConditionTrue(subject.Status.Conditions, userv1.ConditionEmailSent))
	condition := apimeta.FindStatusCondition(subject.Status.Conditions, userv1.ConditionEmailSent)
	require.NotNil(t, condition)
	require.Equal(t, ReasonSendFailed, condition.Reason)
}

func Test_InvitationEmailReconciler_Reconcile_NoEmail_Success(t *testing.T) {
	ctx := context.Background()

	subject := baseInvitation()
	subject.Spec.Email = ""

	c := prepareTest(t, subject)

	r := invitationEmailReconciler(c)
	_, err := r.Reconcile(ctx, requestFor(subject))
	require.NoError(t, err)

	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(subject), subject))
	require.Nil(t, apimeta.FindStatusCondition(subject.Status.Conditions, userv1.ConditionEmailSent))
}

func invitationEmailReconciler(c client.WithWatch) *InvitationEmailReconciler {
	return &InvitationEmailReconciler{
		Client:         c,
		Scheme:         c.Scheme(),
		Recorder:       record.NewFakeRecorder(3),
		MailSender:     &SenderWithConstantId{},
		BaseRetryDelay: time.Minute,
	}
}

func invitationEmailReconcilerWithFailingSender(c client.WithWatch) *InvitationEmailReconciler {
	return &InvitationEmailReconciler{
		Client:         c,
		Scheme:         c.Scheme(),
		Recorder:       record.NewFakeRecorder(3),
		MailSender:     &FailingSender{},
		BaseRetryDelay: time.Minute,
	}
}

func baseInvitation() *userv1.Invitation {
	return &userv1.Invitation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
		Spec: userv1.InvitationSpec{
			Email: "subject@example.com",
		},
		Status: userv1.InvitationStatus{
			Token: "abc",
		},
	}
}
