package secretstorage_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/appuio/control-api/apiserver/secretstorage"
	"github.com/appuio/control-api/apiserver/testresource"
)

func TestRoundtrip(t *testing.T) {
	c := buildClient(t)
	s, err := secretstorage.NewStorage(new(testresource.TestResource), c, "default")
	require.NoError(t, err)

	w, err := s.Watch(context.Background(), &metainternalversion.ListOptions{})
	require.NoError(t, err)
	defer w.Stop()

	ttr := &testresource.TestResource{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Field1: "test",
	}

	_, err = s.Create(context.Background(), ttr, nil, &metav1.CreateOptions{})
	require.NoError(t, err)

	secret := &corev1.Secret{}
	err = c.Get(context.Background(), client.ObjectKey{Name: "test", Namespace: "default"}, secret)
	require.NoError(t, err)

	invOut, err := s.Get(context.Background(), "test", &metav1.GetOptions{})
	require.NoError(t, err)

	require.Equal(t, ttr.Field1, invOut.(*testresource.TestResource).Field1)

	// Update the object
	ttr.Field1 = "updated"
	_, _, err = s.Update(context.Background(), "test", rest.DefaultUpdatedObjectInfo(ttr), nil, nil, false, &metav1.UpdateOptions{})
	require.NoError(t, err)

	list, err := s.List(context.Background(), &metainternalversion.ListOptions{})
	require.NoError(t, err)
	require.Len(t, list.(*testresource.TestResourceList).Items, 1)
	require.Equal(t, "updated", list.(*testresource.TestResourceList).Items[0].Field1)

	// Delete the object
	_, _, err = s.Delete(context.Background(), "test", nil, &metav1.DeleteOptions{})
	require.NoError(t, err)
	list, err = s.List(context.Background(), &metainternalversion.ListOptions{})
	require.NoError(t, err)
	require.Len(t, list.(*testresource.TestResourceList).Items, 0)

	require.Eventually(t, func() bool {
		select {
		case event := <-w.ResultChan():
			return event.Type == watch.Deleted
		default:
			return false
		}
	}, 200*time.Millisecond, time.Microsecond)
}

func TestStatusSubresource(t *testing.T) {
	c := buildClient(t)
	s, err := secretstorage.NewStorage(new(testresource.TestResourceWithStatus), c, "default")
	require.NoError(t, err)

	ttr := &testresource.TestResourceWithStatus{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Field1: "test",
		Status: testresource.TestResourceWithStatusStatus{
			Num: 7,
		},
	}

	_, err = s.Create(context.Background(), ttr, nil, &metav1.CreateOptions{})
	require.NoError(t, err)
	// Empty status on create, can only be set via the status subresource
	require.Equal(t, 0, ttr.SecretStorageGetStatus().(*testresource.TestResourceWithStatusStatus).Num)
	testStatusValue(t, s, 0)

	// Update the status
	ttr.Status.Num = 42
	_, _, err = s.Update(
		request.WithRequestInfo(request.NewContext(), &request.RequestInfo{Subresource: "status"}),
		"test", rest.DefaultUpdatedObjectInfo(ttr), nil, nil, false, &metav1.UpdateOptions{})
	require.NoError(t, err)
	testStatusValue(t, s, 42)

	// Non status update should not change the status
	ttr.Status.Num = 12
	_, _, err = s.Update(
		request.WithRequestInfo(request.NewContext(), &request.RequestInfo{}),
		"test", rest.DefaultUpdatedObjectInfo(ttr), nil, nil, false, &metav1.UpdateOptions{})
	require.NoError(t, err)
	testStatusValue(t, s, 42)
}

func testStatusValue(t *testing.T, s rest.Getter, expected int) {
	t.Helper()

	ttr, err := s.Get(context.Background(), "test", &metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, expected, ttr.(*testresource.TestResourceWithStatus).SecretStorageGetStatus().(*testresource.TestResourceWithStatusStatus).Num)
}

func buildClient(t *testing.T, initObjs ...client.Object) client.WithWatch {
	scheme := runtime.NewScheme()
	require.NoError(t, testresource.AddToScheme(scheme))
	require.NoError(t, clientgoscheme.AddToScheme(scheme))

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(initObjs...).
		Build()
}
