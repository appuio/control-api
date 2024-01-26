package controllers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/appuio/control-api/controllers"
)

func Test_BillingEntityRBACCronJob_Run(t *testing.T) {
	ctx := context.Background()

	be := baseBillingEntity()
	c := prepareTest(t, be)

	subject := &BillingEntityRBACCronJob{
		Client: c,
	}

	require.NoError(t, subject.Run(ctx))

	var adminRole rbacv1.ClusterRole
	var adminRoleBinding rbacv1.ClusterRoleBinding
	adminRoleName := fmt.Sprintf("billingentities-%s-admin", be.Name)
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: adminRoleName}, &adminRole), "admin role should be created")
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: adminRoleName}, &adminRoleBinding), "admin role binding should be created")

	var viewerRole rbacv1.ClusterRole
	var viewerRoleBinding rbacv1.ClusterRoleBinding
	viewerRoleName := fmt.Sprintf("billingentities-%s-viewer", be.Name)
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: viewerRoleName}, &viewerRole), "viewer role should be created")
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: viewerRoleName}, &viewerRoleBinding), "viewer role binding should be created")

	testSubjects := []rbacv1.Subject{
		{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "User",
			Name:     "testuser",
		},
	}
	viewerRoleBinding.Subjects = testSubjects
	require.NoError(t, c.Update(ctx, &viewerRoleBinding))

	require.NoError(t, subject.Run(ctx))

	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: viewerRoleName}, &viewerRoleBinding))
	require.Equal(t, testSubjects, viewerRoleBinding.Subjects, "role bindings should not be changed")
}
