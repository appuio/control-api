package controllers_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	userv1 "github.com/appuio/control-api/apis/user/v1"
	. "github.com/appuio/control-api/controllers"
)

type FailingBESender struct{}
type BESenderWithConstantId struct{}

func (f *FailingBESender) Send(context.Context, string, any) (string, error) {
	return "", errors.New("Err0r")
}

func (s *BESenderWithConstantId) Send(context.Context, string, any) (string, error) {
	return "ID10", nil
}

func Test_BillingEntityEmailCronJob_Sending_Success(t *testing.T) {
	ctx := context.Background()

	subject := baseBillingEntity()

	c := prepareTest(t, subject)

	j := billingEntityCronJob(c)

	err := j.Run(ctx)
	require.NoError(t, err)

	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(subject), subject))
	require.True(t, apimeta.IsStatusConditionTrue(subject.Status.Conditions, userv1.ConditionEmailSent))
	condition := apimeta.FindStatusCondition(subject.Status.Conditions, userv1.ConditionEmailSent)
	require.Equal(t, "Message ID: ID10", condition.Message)
}

func Test_BillingEntityEmailCronJob_Sending_Failure(t *testing.T) {
	ctx := context.Background()

	subject := baseBillingEntity()

	c := prepareTest(t, subject)

	j := billingEntityCronJobWithFailingSender(c)

	err := j.Run(ctx)
	require.NoError(t, err)

	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(subject), subject))
	require.False(t, apimeta.IsStatusConditionTrue(subject.Status.Conditions, userv1.ConditionEmailSent))
	condition := apimeta.FindStatusCondition(subject.Status.Conditions, userv1.ConditionEmailSent)
	require.NotNil(t, condition)
	require.Equal(t, userv1.ConditionReasonSendFailed, condition.Reason)
}

func billingEntityCronJob(c client.WithWatch) *BillingEntityEmailCronJob {
	r := NewBillingEntityEmailCronJob(
		c,
		record.NewFakeRecorder(3),
		c.Scheme(),
		&SenderWithConstantId{},
		"foo@example.com",
	)
	return &r
}
func Test_BillingEntityEmailCronJob_MetricsCorrect(t *testing.T) {
	ctx := context.Background()

	subject := baseBillingEntity()

	c := prepareTest(t, subject)

	j := billingEntityCronJob(c)

	err := j.Run(ctx)
	require.NoError(t, err)

	reg := prometheus.NewRegistry()
	reg.MustRegister(j.GetMetrics())
	require.NoError(t, testutil.CollectAndCompare(reg, strings.NewReader(`
# HELP control_api_billingentity_emails_sent_failed_total Total number of e-mails which failed to send
# TYPE control_api_billingentity_emails_sent_failed_total counter
control_api_billingentity_emails_sent_failed_total 0
# HELP control_api_billingentity_emails_sent_success_total Total number of successfully sent e-mails
# TYPE control_api_billingentity_emails_sent_success_total counter
control_api_billingentity_emails_sent_success_total 1
`),
	))
}

func Test_BillingEntityEmailCronJob_WithSendingFailure_MetricsCorrect(t *testing.T) {
	ctx := context.Background()

	subject := baseBillingEntity()

	c := prepareTest(t, subject)

	j := billingEntityCronJobWithFailingSender(c)

	err := j.Run(ctx)
	require.NoError(t, err)

	reg := prometheus.NewRegistry()
	reg.MustRegister(j.GetMetrics())
	require.NoError(t, testutil.CollectAndCompare(reg, strings.NewReader(`
# HELP control_api_billingentity_emails_sent_failed_total Total number of e-mails which failed to send
# TYPE control_api_billingentity_emails_sent_failed_total counter
control_api_billingentity_emails_sent_failed_total 1
# HELP control_api_billingentity_emails_sent_success_total Total number of successfully sent e-mails
# TYPE control_api_billingentity_emails_sent_success_total counter
control_api_billingentity_emails_sent_success_total 0
`),
	))
}

func billingEntityCronJobWithFailingSender(c client.WithWatch) *BillingEntityEmailCronJob {
	r := NewBillingEntityEmailCronJob(
		c,
		record.NewFakeRecorder(3),
		c.Scheme(),
		&FailingSender{},
		"foo@example.com",
	)
	return &r
}

func baseBillingEntity() *billingv1.BillingEntity {
	return &billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name: "be-111",
		},
		Spec: billingv1.BillingEntitySpec{
			Name: "myCompany",
		},
		Status: billingv1.BillingEntityStatus{
			Conditions: []metav1.Condition{},
		},
	}
}
