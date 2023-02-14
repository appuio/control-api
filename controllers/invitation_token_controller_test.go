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

	userv1 "github.com/appuio/control-api/apis/user/v1"
	. "github.com/appuio/control-api/controllers"
)

func Test_InvitationTokenReconciler_Reconcile_Success(t *testing.T) {
	const tokenValidFor = time.Minute
	ctx := context.Background()

	subject := userv1.Invitation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
	}

	c := prepareTest(t, &subject)

	_, err := (&InvitationTokenReconciler{
		Client:   c,
		Scheme:   c.Scheme(),
		Recorder: record.NewFakeRecorder(3),

		TokenValidFor: tokenValidFor,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})
	require.NoError(t, err)

	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: subject.Name}, &subject))
	assert.NotEmpty(t, subject.Status.Token)
	assert.WithinDuration(t, time.Now().Add(tokenValidFor), subject.Status.ValidUntil.Time, time.Second)
}
