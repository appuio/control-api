package controllers_test

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	"github.com/appuio/control-api/controllers"
)

func TestEmailPendingMetric(t *testing.T) {
	c := prepareTest(t, &userv1.Invitation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-inv",
		},
		Spec:   userv1.InvitationSpec{},
		Status: userv1.InvitationStatus{Conditions: []metav1.Condition{{Type: userv1.ConditionEmailSent, Status: metav1.ConditionTrue}}},
	}, &userv1.Invitation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-inv-2",
		},
		Spec: userv1.InvitationSpec{},
	})

	require.NoError(t,
		testutil.CollectAndCompare(&controllers.EmailPendingMetric{c}, strings.NewReader(`
# HELP control_api_invitation_emails_pending_current Amount of e-mails that have not been sent yet
# TYPE control_api_invitation_emails_pending_current gauge
control_api_invitation_emails_pending_current 1
`),
		),
	)
}
