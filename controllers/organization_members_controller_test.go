package controllers_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	controlv1 "github.com/appuio/control-api/apis/v1"
	. "github.com/appuio/control-api/controllers"
)

var testMemb = controlv1.OrganizationMembers{
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

var testUserPrefix = "control-api#"

func Test_OrganizationMembersReconciler_Reconcile_Sucess(t *testing.T) {
	ctx := context.Background()
	c := prepareTest(t, &testMemb)
	fakeRecorder := record.NewFakeRecorder(3)
	membRoles := []string{"foo", "bar", "buzz"}

	_, err := (&OrganizationMembersReconciler{
		Client:   c,
		Scheme:   c.Scheme(),
		Recorder: fakeRecorder,

		MemberRoles: membRoles,
		UserPrefix:  testUserPrefix,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      testMemb.Name,
			Namespace: testMemb.Namespace,
		},
	})
	require.NoError(t, err)

	for _, role := range membRoles {
		testRoleExists(t, c, role, testUserPrefix, testMemb)
	}
}

func Test_OrganizationMembersReconciler_Reconcile_PartialFailure(t *testing.T) {
	ctx := context.Background()
	c := failingClient{prepareTest(t, &testMemb)}
	fakeRecorder := record.NewFakeRecorder(3)
	membRoles := []string{"foo", "fail-bar", "buzz"}

	_, err := (&OrganizationMembersReconciler{
		Client:   c,
		Scheme:   c.Scheme(),
		Recorder: fakeRecorder,

		MemberRoles: membRoles,
		UserPrefix:  testUserPrefix,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      testMemb.Name,
			Namespace: testMemb.Namespace,
		},
	})
	assert.Error(t, err)

	for _, role := range []string{"foo", "buzz"} {
		testRoleExists(t, c, role, testUserPrefix, testMemb)
	}
	rb := rbacv1.RoleBinding{}
	assert.Errorf(t, c.Get(ctx, types.NamespacedName{Name: "fail-bar", Namespace: testMemb.Namespace}, &rb), "don't create bar")
	require.Len(t, fakeRecorder.Events, 1)
}

func Test_OrganizationMembersReconciler_Reconcile_MissmatchStatus(t *testing.T) {
	ctx := context.Background()
	membRoles := []string{"foo", "bar", "buzz"}
	memb := testMemb
	memb.Status.ResolvedUserRefs = []controlv1.UserRef{
		{Name: "u1"},
		{Name: "u3"},
	}
	memb.Spec.UserRefs = []controlv1.UserRef{
		{Name: "u2"},
		{Name: "u3"},
	}

	c := prepareTest(t, &memb)
	fakeRecorder := record.NewFakeRecorder(3)

	_, err := (&OrganizationMembersReconciler{
		Client:   c,
		Scheme:   c.Scheme(),
		Recorder: fakeRecorder,

		MemberRoles: membRoles,
		UserPrefix:  testUserPrefix,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      memb.Name,
			Namespace: memb.Namespace,
		},
	})
	assert.NoError(t, err)

	for _, role := range []string{"foo", "bar", "buzz"} {
		testRoleExists(t, c, role, testUserPrefix, memb)
	}
}

func testRoleExists(t *testing.T, c client.WithWatch, role, userPrefix string, memb controlv1.OrganizationMembers) {
	t.Run(role+" exists", func(t *testing.T) {
		rb := rbacv1.RoleBinding{}
		require.NoError(t, c.Get(context.TODO(), types.NamespacedName{Name: role, Namespace: memb.Namespace}, &rb))

		assert.Equal(t, rb.RoleRef, rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     role,
		})
		users := memb.Spec.UserRefs
		require.Len(t, rb.Subjects, len(users))
		for _, u := range users {
			assert.Contains(t, rb.Subjects, rbacv1.Subject{
				APIGroup: rbacv1.GroupName,
				Kind:     "User",
				Name:     userPrefix + u.Name,
			})
		}

		require.Len(t, rb.OwnerReferences, 1, "controller must set owner reference")
		assert.Equal(t, memb.Name, rb.OwnerReferences[0].Name, "owner reference name must match")
	})
}

type failingClient struct {
	client.WithWatch
}

func (c failingClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if strings.HasPrefix(obj.GetName(), "fail-") {
		return apierrors.NewInternalError(errors.New("ups"))
	}
	return c.WithWatch.Create(ctx, obj, opts...)
}
