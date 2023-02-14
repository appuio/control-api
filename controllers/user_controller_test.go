package controllers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	controlv1 "github.com/appuio/control-api/apis/v1"
	. "github.com/appuio/control-api/controllers"
)

func Test_UserController_Reconcile_Success(t *testing.T) {
	ctx := context.Background()
	userPrefix := "appuio#"
	rolePrefix := "control-api:user:"

	subject := controlv1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
	}

	fakeRecorder := record.NewFakeRecorder(3)

	c := prepareTest(t, &subject)

	_, err := (&UserReconciler{
		Client:   c,
		Scheme:   c.Scheme(),
		Recorder: fakeRecorder,

		UserPrefix: userPrefix,
		RolePrefix: rolePrefix,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})
	require.NoError(t, err)

	cr := rbacv1.ClusterRole{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: rolePrefix + "subject-owner"}, &cr))
	assert.Len(t, cr.OwnerReferences, 1, "controller must set owner reference")
	assert.Equal(t, subject.Name, cr.OwnerReferences[0].Name, "owner reference name must match")
	assert.Len(t, cr.Rules, 1, "ClusterRole should have rule referencing the user")
	assert.Equal(t, []string{subject.Name}, cr.Rules[0].ResourceNames, "rule resource name must match")

	crb := rbacv1.ClusterRoleBinding{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: rolePrefix + "subject-owner"}, &crb))
	assert.Len(t, crb.OwnerReferences, 1, "controller must set owner reference")
	assert.Equal(t, subject.Name, crb.OwnerReferences[0].Name, "owner reference name must match")
	assert.Len(t, crb.Subjects, 1, "ClusterRoleBinding should have subject referencing the user")
	assert.Equal(t, userPrefix+subject.Name, crb.Subjects[0].Name, "user name must match and include prefix")
}
