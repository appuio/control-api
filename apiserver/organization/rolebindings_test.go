package organization

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
)

func TestGenerateRoleBinding(t *testing.T) {
	rb, err := generateRoleBinding(request.WithUser(request.NewContext(), &user.DefaultInfo{
		Name: "fooUser",
	}), "foobar", "fooRole")
	require.NoError(t, err)
	require.Len(t, rb.Subjects, 1)
	assert.Equal(t, rbacv1.Subject{
		Kind:     "User",
		APIGroup: rbacv1.GroupName,
		Name:     "fooUser",
	}, rb.Subjects[0])
	assert.Equal(t, "foobar", rb.Namespace)
	assert.Equal(t, rbacv1.RoleRef{
		APIGroup: rbacv1.GroupName,
		Kind:     "ClusterRole",
		Name:     "fooRole",
	}, rb.RoleRef)
}
