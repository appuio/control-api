package controllers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"
	. "github.com/appuio/control-api/controllers"
)

func Test_UserController_Reconcile_Success(t *testing.T) {
	ctx := context.Background()
	userPrefix := "appuio#"

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
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})
	require.NoError(t, err)

	cr := rbacv1.ClusterRole{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: "subject-owner"}, &cr))
	assert.Len(t, cr.OwnerReferences, 1, "controller must set owner reference")
	assert.Equal(t, subject.Name, cr.OwnerReferences[0].Name, "owner reference name must match")
	assert.Len(t, cr.Rules, 1, "ClusterRole should have rule referencing the user")
	assert.Equal(t, []string{subject.Name}, cr.Rules[0].ResourceNames, "rule resource name must match")

	crb := rbacv1.ClusterRoleBinding{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: "subject-owner"}, &crb))
	assert.Len(t, crb.OwnerReferences, 1, "controller must set owner reference")
	assert.Equal(t, subject.Name, crb.OwnerReferences[0].Name, "owner reference name must match")
	assert.Len(t, crb.Subjects, 1, "ClusterRoleBinding should have subject referencing the user")
	assert.Equal(t, userPrefix+subject.Name, crb.Subjects[0].Name, "user name must match and include prefix")
}

func prepareTest(t *testing.T, initObjs ...client.Object) client.WithWatch {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(orgv1.AddToScheme(scheme))
	utilruntime.Must(controlv1.AddToScheme(scheme))

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(initObjs...).
		Build()
}

func boolP(b bool) *bool {
	return &b
}
