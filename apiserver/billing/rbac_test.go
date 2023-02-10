package billing

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/appuio/control-api/apiserver/authwrapper/mock"
	"github.com/appuio/control-api/apiserver/testresource"
)

func Test_createRBACWrapper(t *testing.T) {
	user := "testuser"
	returnedResourceName := "be-2345"

	c := newClient()
	ctrl, store := newStore(t)
	defer ctrl.Finish()

	subject := &createRBACWrapper{
		Storage: clusterScopedStorage{store},
		client:  c,
	}

	store.EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&testresource.TestResource{ObjectMeta: metav1.ObjectMeta{Name: returnedResourceName}}, nil).
		Times(1)

	_, err := subject.Create(ctxWithInfo("create", "", user), &testresource.TestResource{}, nil, &metav1.CreateOptions{})
	require.NoError(t, err)

	rn := "billingentities-" + returnedResourceName + "-viewer"
	var role rbacv1.ClusterRole
	require.NoError(t, c.Get(context.Background(), types.NamespacedName{Name: rn}, &role))
	var clusterrole rbacv1.ClusterRoleBinding
	require.NoError(t, c.Get(context.Background(), types.NamespacedName{Name: rn}, &clusterrole))
	assert.Equal(t, []rbacv1.Subject{{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: user}}, clusterrole.Subjects)
}

func Test_createRBACWrapper_rollback(t *testing.T) {
	user := "testuser"
	returnedResourceName := "be-2345"

	c := newClient()
	ctrl, store := newStore(t)
	defer ctrl.Finish()

	subject := &createRBACWrapper{
		Storage: clusterScopedStorage{store},
		client:  c,
	}

	store.EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&testresource.TestResource{ObjectMeta: metav1.ObjectMeta{Name: returnedResourceName}}, nil).
		Times(1)

	store.EXPECT().
		Delete(gomock.Any(), returnedResourceName, gomock.Any(), gomock.Any()).
		Return(nil, true, nil).
		Times(1)

	// Force an already exist error to trigger rollback
	rn := "billingentities-" + returnedResourceName + "-viewer"
	require.NoError(t,
		c.Create(context.Background(), &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: rn}}),
	)

	_, err := subject.Create(ctxWithInfo("create", "", user), &testresource.TestResource{}, nil, &metav1.CreateOptions{})
	require.Error(t, err, "expected error on create to trigger rollback")

	var role rbacv1.ClusterRole
	err = c.Get(context.Background(), types.NamespacedName{Name: rn}, &role)
	assert.True(t, apierrors.IsNotFound(err), "expected role to be deleted on rollback")
}

func newStore(t *testing.T) (*gomock.Controller, *mock.MockStandardStorage) {
	ctrl := gomock.NewController(t)
	store := mock.NewMockStandardStorage(ctrl)
	return ctrl, store
}

func newClient() client.WithWatch {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	return fake.NewClientBuilder().WithScheme(scheme).Build()
}

type clusterScopedStorage struct {
	rest.StandardStorage
}

func (clusterScopedStorage) NamespaceScoped() bool {
	return false
}

func ctxWithInfo(verb string, objName string, username string) context.Context {
	gvr := (&testresource.TestResource{}).GetGroupVersionResource()
	return request.WithUser(
		request.WithRequestInfo(request.NewContext(),
			&request.RequestInfo{
				APIGroup:   gvr.Group,
				APIVersion: gvr.Version,
				Resource:   gvr.Resource,

				Verb: verb,
				Name: objName,
			}),
		&user.DefaultInfo{
			Name: username,
		})
}
