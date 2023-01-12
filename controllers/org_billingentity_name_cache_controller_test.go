package controllers_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	. "github.com/appuio/control-api/controllers"
)

func Test_OrgBillingEntityNameCacheController_Reconcile_Success(t *testing.T) {
	ctx := context.Background()
	subject := orgv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
		Spec: orgv1.OrganizationSpec{
			BillingEntityRef: "be",
		},
	}

	c := prepareTest(t, &subject, &billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name: "be",
		},
		Spec: billingv1.BillingEntitySpec{
			Name: "be-name",
		},
	})

	res, err := (&OrgBillingEntityNameCacheController{
		Client:   c,
		Scheme:   c.Scheme(),
		Recorder: record.NewFakeRecorder(3),

		RefreshInterval: time.Minute,
		RefreshJitter:   time.Second,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, res.RequeueAfter, time.Minute)

	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: subject.Name}, &subject))
	assert.Equal(t, "be-name", subject.Status.BillingEntityName)
}
