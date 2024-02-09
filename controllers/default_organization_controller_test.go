package controllers_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	controlv1 "github.com/appuio/control-api/apis/v1"
	. "github.com/appuio/control-api/controllers"
)

var testMemberships1 = controlv1.OrganizationMembers{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "members",
		Namespace: "foo-gmbh",
	},
	Spec: controlv1.OrganizationMembersSpec{
		UserRefs: []controlv1.UserRef{
			{Name: "u1"},
			{Name: "u2"},
			{Name: "u3"},
		},
	},
	Status: controlv1.OrganizationMembersStatus{
		ResolvedUserRefs: []controlv1.UserRef{
			{Name: "u1"},
			{Name: "u2"},
			{Name: "u3"},
		},
	},
}

var testMemberships2 = controlv1.OrganizationMembers{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "members",
		Namespace: "bar-gmbh",
	},
	Spec: controlv1.OrganizationMembersSpec{
		UserRefs: []controlv1.UserRef{
			{Name: "u1"},
		},
	},
	Status: controlv1.OrganizationMembersStatus{
		ResolvedUserRefs: []controlv1.UserRef{
			{Name: "u1"},
		},
	},
}

var u1 = controlv1.User{
	ObjectMeta: metav1.ObjectMeta{
		Name: "u1",
	},
}
var u2 = controlv1.User{
	ObjectMeta: metav1.ObjectMeta{
		Name: "u2",
	},
}
var u3 = controlv1.User{
	ObjectMeta: metav1.ObjectMeta{
		Name: "u3",
	},
}

func Test_DefaultOrganizationReconciler_Reconcile_Success(t *testing.T) {
	ctx := context.Background()
	c := prepareTest(t, &testMemberships1, &testMemberships2, &u1, &u2, &u3)
	fakeRecorder := record.NewFakeRecorder(3)

	_, err := (&DefaultOrganizationReconciler{
		Client:   c,
		Scheme:   c.Scheme(),
		Recorder: fakeRecorder,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      testMemberships1.Name,
			Namespace: testMemberships1.Namespace,
		},
	})
	require.NoError(t, err)

	user := controlv1.User{}
	require.NoError(t, c.Get(context.TODO(), types.NamespacedName{Name: u1.ObjectMeta.Name}, &user))
	assert.Empty(t, user.Spec.Preferences.DefaultOrganizationRef)

	require.NoError(t, c.Get(context.TODO(), types.NamespacedName{Name: u2.ObjectMeta.Name}, &user))
	assert.Equal(t, testMemberships1.ObjectMeta.Namespace, user.Spec.Preferences.DefaultOrganizationRef)

	require.NoError(t, c.Get(context.TODO(), types.NamespacedName{Name: u3.ObjectMeta.Name}, &user))
	assert.Equal(t, testMemberships1.ObjectMeta.Namespace, user.Spec.Preferences.DefaultOrganizationRef)

}

func Test_DefaultOrganizationReconciler_Reconcile_NoMembership_Success(t *testing.T) {
	ctx := context.Background()
	c := prepareTest(t, &testMemberships2, &u1, &u2, &u3)
	fakeRecorder := record.NewFakeRecorder(3)

	_, err := (&DefaultOrganizationReconciler{
		Client:   c,
		Scheme:   c.Scheme(),
		Recorder: fakeRecorder,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      testMemberships2.Name,
			Namespace: testMemberships2.Namespace,
		},
	})
	require.NoError(t, err)

	user := controlv1.User{}
	require.NoError(t, c.Get(context.TODO(), types.NamespacedName{Name: u1.ObjectMeta.Name}, &user))
	assert.Equal(t, testMemberships2.ObjectMeta.Namespace, user.Spec.Preferences.DefaultOrganizationRef)

	require.NoError(t, c.Get(context.TODO(), types.NamespacedName{Name: u2.ObjectMeta.Name}, &user))
	assert.Empty(t, user.Spec.Preferences.DefaultOrganizationRef)

	require.NoError(t, c.Get(context.TODO(), types.NamespacedName{Name: u3.ObjectMeta.Name}, &user))
	assert.Empty(t, user.Spec.Preferences.DefaultOrganizationRef)

}

func Test_DefaultOrganizationReconciler_Reconcile_UserNotExist_Success(t *testing.T) {
	ctx := context.Background()
	c := prepareTest(t, &testMemberships1, &u1)
	fakeRecorder := record.NewFakeRecorder(3)

	_, err := (&DefaultOrganizationReconciler{
		Client:   c,
		Scheme:   c.Scheme(),
		Recorder: fakeRecorder,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      testMemberships1.Name,
			Namespace: testMemberships1.Namespace,
		},
	})
	require.NoError(t, err)

	user := controlv1.User{}
	require.NoError(t, c.Get(context.TODO(), types.NamespacedName{Name: u1.ObjectMeta.Name}, &user))
	assert.Equal(t, testMemberships1.ObjectMeta.Namespace, user.Spec.Preferences.DefaultOrganizationRef)

	require.NoError(t, c.Get(context.TODO(), types.NamespacedName{Name: u2.ObjectMeta.Name}, &user))
	assert.Equal(t, testMemberships1.ObjectMeta.Namespace, user.Spec.Preferences.DefaultOrganizationRef)

	require.NoError(t, c.Get(context.TODO(), types.NamespacedName{Name: u3.ObjectMeta.Name}, &user))
	assert.Equal(t, testMemberships1.ObjectMeta.Namespace, user.Spec.Preferences.DefaultOrganizationRef)

}

func Test_DefaultOrganizationReconciler_Reconcile_Error(t *testing.T) {
	failU4 := controlv1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fail-u4",
		},
	}
	failMemberships := controlv1.OrganizationMembers{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "members",
			Namespace: "foo-gmbh",
		},
		Spec: controlv1.OrganizationMembersSpec{
			UserRefs: []controlv1.UserRef{
				{Name: "u1"},
				{Name: "u2"},
				{Name: "u3"},
				{Name: "fail-u4"},
			},
		},
		Status: controlv1.OrganizationMembersStatus{
			ResolvedUserRefs: []controlv1.UserRef{
				{Name: "u1"},
				{Name: "u2"},
				{Name: "u3"},
				{Name: "fail-u4"},
			},
		},
	}
	ctx := context.Background()
	c := failingClient{prepareTest(t, &failMemberships, &failU4, &u1, &u2, &u3)}
	fakeRecorder := record.NewFakeRecorder(3)

	_, err := (&DefaultOrganizationReconciler{
		Client:   c,
		Scheme:   c.Scheme(),
		Recorder: fakeRecorder,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      failMemberships.Name,
			Namespace: failMemberships.Namespace,
		},
	})
	require.Error(t, err)

	user := controlv1.User{}
	require.NoError(t, c.Get(context.TODO(), types.NamespacedName{Name: u1.ObjectMeta.Name}, &user))
	assert.Equal(t, failMemberships.ObjectMeta.Namespace, user.Spec.Preferences.DefaultOrganizationRef)

	require.NoError(t, c.Get(context.TODO(), types.NamespacedName{Name: u2.ObjectMeta.Name}, &user))
	assert.Equal(t, failMemberships.ObjectMeta.Namespace, user.Spec.Preferences.DefaultOrganizationRef)

	require.NoError(t, c.Get(context.TODO(), types.NamespacedName{Name: u3.ObjectMeta.Name}, &user))
	assert.Equal(t, failMemberships.ObjectMeta.Namespace, user.Spec.Preferences.DefaultOrganizationRef)

}

func (c failingClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if strings.HasPrefix(obj.GetName(), "fail-") {
		return apierrors.NewInternalError(errors.New("ups"))
	}
	return c.WithWatch.Update(ctx, obj, opts...)
}
