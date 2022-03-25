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

func Test_OrganizationMembersReconciler_Reconcile_Sucess(t *testing.T) {
	ctx := context.Background()
	users := []controlv1.UserRef{
		{Name: "u1"},
		{Name: "u2"},
		{Name: "u3"},
	}
	memb := controlv1.OrganizationMembers{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "members",
			Namespace: "foo-gmbh",
		},
		Spec: controlv1.OrganizationMembersSpec{
			UserRefs: users,
		},
		Status: controlv1.OrganizationMembersStatus{
			ResolvedUserRefs: users,
		},
	}
	c := prepareTest(t, &memb)
	fakeRecorder := record.NewFakeRecorder(3)
	membRoles := []string{"foo", "bar", "buzz"}
	userPrefix := "control-api#"

	_, err := (&OrganizationMembersReconciler{
		Client:   c,
		Scheme:   c.Scheme(),
		Recorder: fakeRecorder,

		MemberRoles: membRoles,
		UserPrefix:  userPrefix,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      memb.Name,
			Namespace: memb.Namespace,
		},
	})
	require.NoError(t, err)

	for _, role := range membRoles {
		t.Run("create "+role, func(t *testing.T) {
			rb := rbacv1.RoleBinding{}
			require.NoError(t, c.Get(ctx, types.NamespacedName{Name: role, Namespace: memb.Namespace}, &rb))

			assert.Equal(t, rb.RoleRef, rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "ClusterRole",
				Name:     role,
			})
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
}
