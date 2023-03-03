package targetref_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"
	"github.com/appuio/control-api/controllers/targetref"
)

func Test_GetTarget(t *testing.T) {
	tt := []client.Object{
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
		},
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
		},
		&controlv1.OrganizationMembers{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
		},
		&controlv1.Team{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
		},
	}

	for _, obj := range tt {
		t.Run(fmt.Sprintf("%T", obj), func(t *testing.T) {
			c := newClient(t)
			require.NoError(t, c.Create(context.Background(), obj))
			// Fill type meta
			require.NoError(t, c.Get(context.Background(), client.ObjectKeyFromObject(obj), obj))

			getObj, err := targetref.GetTarget(context.Background(), c, userv1.TargetRef{
				APIGroup:  obj.GetObjectKind().GroupVersionKind().Group,
				Kind:      obj.GetObjectKind().GroupVersionKind().Kind,
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			})
			require.NoError(t, err)
			require.IsType(t, obj, getObj)
		})
	}
}

func Test_GetTarget_UnsupportedType(t *testing.T) {
	_, err := targetref.GetTarget(context.Background(), nil, userv1.TargetRef{
		APIGroup: "apps/v1",
		Kind:     "Deployment",
	})
	require.ErrorContains(t, err, "unsupported target")
	require.ErrorContains(t, err, "Deployment")
}

func Test_UserAccessor_AccessUser(t *testing.T) {
	tt := []client.Object{
		&rbacv1.ClusterRoleBinding{
			Subjects: []rbacv1.Subject{
				{
					Kind:     rbacv1.UserKind,
					APIGroup: rbacv1.GroupName,
					Name:     "user1",
				},
			},
		},
		&rbacv1.RoleBinding{
			Subjects: []rbacv1.Subject{
				{
					Kind:     rbacv1.UserKind,
					APIGroup: rbacv1.GroupName,
					Name:     "user1",
				},
			},
		},
		&controlv1.OrganizationMembers{
			Spec: controlv1.OrganizationMembersSpec{
				UserRefs: []controlv1.UserRef{
					{Name: "user1"},
				},
			},
		},
		&controlv1.Team{
			Spec: controlv1.TeamSpec{
				UserRefs: []controlv1.UserRef{
					{Name: "user1"},
				},
			},
		},
	}

	for _, obj := range tt {
		t.Run(fmt.Sprintf("%T", obj), func(t *testing.T) {
			a, err := targetref.NewUserAccessor(obj)
			require.NoError(t, err)

			require.False(t, a.HasUser("user2"))
			require.True(t, a.EnsureUser("user2"))
			require.True(t, a.HasUser("user2"))
			require.False(t, a.EnsureUser("user2"))
			require.True(t, a.HasUser("user2"))
		})
	}
}

func Test_UserAccessor_UnsupportedType(t *testing.T) {
	_, err := targetref.NewUserAccessor(&rbacv1.Role{})
	require.Error(t, err)
}

func newClient(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()

	scheme := runtime.NewScheme()
	require.NoError(t, rbacv1.AddToScheme(scheme))
	require.NoError(t, clientgoscheme.AddToScheme(scheme))
	require.NoError(t, controlv1.AddToScheme(scheme))
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		Build()
}
