package authwrapper_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"

	"github.com/appuio/control-api/apiserver/authwrapper"
	"github.com/appuio/control-api/apiserver/authwrapper/mock"
	"github.com/appuio/control-api/apiserver/authwrapper/testresource"
)

var gvr = func() metav1.GroupVersionResource {
	gvr := (&testresource.TestResource{}).GetGroupVersionResource()
	return metav1.GroupVersionResource{
		Group:    gvr.Group,
		Version:  gvr.Version,
		Resource: gvr.Resource,
	}
}()

func TestGet(t *testing.T) {
	ctrl, store, mauth, subject := setupStandardStorage(t)
	defer ctrl.Finish()

	t.Run("allow", func(t *testing.T) {
		allowAuthResponse(mauth)
		store.EXPECT().
			Get(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, nil).
			Times(1)
		_, err := subject.Get(ctxWithInfo("get", "tr1"), "tr1", nil)
		assert.NoError(t, err)
	})

	t.Run("deny", func(t *testing.T) {
		denyAuthResponse(mauth)
		_, err := subject.Get(ctxWithInfo("get", "tr1"), "tr1", nil)
		assert.ErrorContains(t, err, "forbidden")
	})

	t.Run("not implemented", func(t *testing.T) {
		allowAuthResponse(mauth)

		basicStore := mock.NewMockStorage(ctrl)
		subject := mustAuthorizedStorage(t, clusterScopedStorage{basicStore}, gvr, mauth).(rest.Getter)
		_, err := subject.Get(ctxWithInfo("get", "tr1"), "tr1", nil)
		assert.ErrorContains(t, err, "not supported")
	})
}

func TestCreate(t *testing.T) {
	ctrl, store, mauth, subject := setupStandardStorage(t)
	defer ctrl.Finish()

	t.Run("allow", func(t *testing.T) {
		allowAuthResponse(mauth)
		store.EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, nil).
			Times(1)
		_, err := subject.Create(ctxWithInfo("create", ""), &testresource.TestResource{}, nil, nil)
		assert.NoError(t, err)
	})

	t.Run("deny", func(t *testing.T) {
		denyAuthResponse(mauth)
		_, err := subject.Create(ctxWithInfo("create", ""), &testresource.TestResource{}, nil, nil)
		assert.ErrorContains(t, err, "forbidden")
	})

	t.Run("not implemented", func(t *testing.T) {
		allowAuthResponse(mauth)

		basicStore := mock.NewMockStorage(ctrl)
		subject := mustAuthorizedStorage(t, clusterScopedStorage{basicStore}, gvr, mauth).(rest.Creater)
		_, err := subject.Create(ctxWithInfo("create", ""), &testresource.TestResource{}, nil, nil)
		assert.ErrorContains(t, err, "not supported")
	})
}

func TestUpdate(t *testing.T) {
	ctrl, store, mauth, subject := setupStandardStorage(t)
	defer ctrl.Finish()

	t.Run("allow", func(t *testing.T) {
		allowAuthResponse(mauth)
		store.EXPECT().
			Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, false, nil).
			Times(1)
		_, _, err := subject.Update(ctxWithInfo("update", "tr1"), "tr1", testUpdateInfoNoUpdate, nil, nil, false, nil)
		assert.NoError(t, err)
	})

	t.Run("deny", func(t *testing.T) {
		denyAuthResponse(mauth)
		_, _, err := subject.Update(ctxWithInfo("update", "tr1"), "tr1", testUpdateInfoNoUpdate, nil, nil, false, nil)
		assert.ErrorContains(t, err, "forbidden")
	})

	t.Run("not implemented", func(t *testing.T) {
		allowAuthResponse(mauth)

		basicStore := mock.NewMockStorage(ctrl)
		subject := mustAuthorizedStorage(t, clusterScopedStorage{basicStore}, gvr, mauth).(rest.Updater)
		_, _, err := subject.Update(ctxWithInfo("update", "tr1"), "tr1", testUpdateInfoNoUpdate, nil, nil, false, nil)
		assert.ErrorContains(t, err, "not supported")
	})
}

func TestDelete(t *testing.T) {
	ctrl, store, mauth, subject := setupStandardStorage(t)
	defer ctrl.Finish()

	t.Run("allow", func(t *testing.T) {
		allowAuthResponse(mauth)
		store.EXPECT().
			Delete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, false, nil).
			Times(1)
		_, _, err := subject.Delete(ctxWithInfo("delete", "tr1"), "tr1", nil, nil)
		assert.NoError(t, err)
	})

	t.Run("deny", func(t *testing.T) {
		denyAuthResponse(mauth)
		_, _, err := subject.Delete(ctxWithInfo("delete", "tr1"), "tr1", nil, nil)
		assert.ErrorContains(t, err, "forbidden")
	})

	t.Run("not implemented", func(t *testing.T) {
		allowAuthResponse(mauth)

		basicStore := mock.NewMockStorage(ctrl)
		subject := mustAuthorizedStorage(t, clusterScopedStorage{basicStore}, gvr, mauth).(rest.GracefulDeleter)
		_, _, err := subject.Delete(ctxWithInfo("delete", "tr1"), "tr1", nil, nil)
		assert.ErrorContains(t, err, "not supported")
	})
}

func TestDeleteCollection(t *testing.T) {
	ctrl, _, _, subject := setupStandardStorage(t)
	defer ctrl.Finish()

	t.Run("not implemented", func(t *testing.T) {
		_, err := subject.DeleteCollection(ctxWithInfo("delete", ""), nil, nil, nil)
		assert.ErrorContains(t, err, "not supported")
	})
}

func TestList(t *testing.T) {
	ctrl, store, mauth, subject := setupStandardStorage(t)
	defer ctrl.Finish()

	t.Run("allow", func(t *testing.T) {
		allowAuthResponse(mauth)
		store.EXPECT().
			NewList().
			Return((&testresource.TestResource{}).NewList()).
			Times(1)
		store.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return((&testresource.TestResource{}).NewList(), nil).
			Times(1)
		_, err := subject.List(ctxWithInfo("list", ""), nil)
		assert.NoError(t, err)
	})

	t.Run("filter list", func(t *testing.T) {
		gomock.InOrder(
			allowAuthResponse(mauth),
			denyAuthResponse(mauth),
			allowAuthResponse(mauth),
		)
		store.EXPECT().
			NewList().
			Return((&testresource.TestResource{}).NewList()).
			Times(1)
		store.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return((&testresource.TestResourceList{
				Items: []testresource.TestResource{
					{ObjectMeta: metav1.ObjectMeta{Name: "tr1"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "tr2"}},
				},
			}), nil).
			Times(1)
		list, err := subject.List(ctxWithInfo("list", ""), nil)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(list.(*testresource.TestResourceList).Items))
	})

	t.Run("deny", func(t *testing.T) {
		denyAuthResponse(mauth)
		_, err := subject.List(ctxWithInfo("list", ""), nil)
		assert.ErrorContains(t, err, "forbidden")
	})

}

func TestWatch(t *testing.T) {
	ctrl, store, mauth, subject := setupStandardStorage(t)
	defer ctrl.Finish()

	t.Run("allow", func(t *testing.T) {
		allowAuthResponse(mauth)
		store.EXPECT().
			Watch(gomock.Any(), gomock.Any()).
			Return(testWatcher{}, nil).
			Times(1)
		_, err := subject.Watch(ctxWithInfo("watch", ""), nil)
		assert.NoError(t, err)
	})

	t.Run("filter watch", func(t *testing.T) {
		gomock.InOrder(
			allowAuthResponse(mauth),
			denyAuthResponse(mauth),
			allowAuthResponse(mauth),
			denyAuthResponse(mauth),
		)

		events := []watch.Event{
			{Type: watch.Error, Object: &metav1.Status{Message: "error"}},               // Passed as is
			{Type: watch.Added, Object: &metav1.Status{Message: "missing object meta"}}, // Dropped
			{Type: watch.Added, Object: &testresource.TestResource{ObjectMeta: metav1.ObjectMeta{Name: "tr1"}}},
			{Type: watch.Added, Object: &testresource.TestResource{ObjectMeta: metav1.ObjectMeta{Name: "tr2"}}},
			{Type: watch.Added, Object: &testresource.TestResource{ObjectMeta: metav1.ObjectMeta{Name: "tr3"}}},
		}
		eventChan := make(chan watch.Event, len(events))
		for _, event := range events {
			eventChan <- event
		}
		close(eventChan)
		tw := testWatcher{eventChan}
		store.EXPECT().
			Watch(gomock.Any(), gomock.Any()).
			Return(tw, nil).
			Times(1)
		filtered, err := subject.Watch(ctxWithInfo("watch", ""), nil)
		assert.NoError(t, err)

		collected := []watch.Event{}
		for {
			event, ok := <-filtered.ResultChan()
			if !ok {
				break
			}
			collected = append(collected, event)
		}
		require.Len(t, collected, 2)
		assert.Equal(t, watch.Error, collected[0].Type)
		assert.Equal(t, "tr2", collected[1].Object.(*testresource.TestResource).Name)
	})

	t.Run("deny", func(t *testing.T) {
		denyAuthResponse(mauth)
		_, err := subject.Watch(ctxWithInfo("watch", ""), nil)
		assert.ErrorContains(t, err, "forbidden")
	})

	t.Run("not implemented", func(t *testing.T) {
		allowAuthResponse(mauth)

		basicStore := mock.NewMockStorage(ctrl)
		subject := mustAuthorizedStorage(t, clusterScopedStorage{storageWithList{basicStore}}, gvr, mauth).(rest.Watcher)
		_, err := subject.Watch(ctxWithInfo("watch", ""), nil)
		assert.ErrorContains(t, err, "not supported")
	})
}

func setupStandardStorage(t *testing.T) (*gomock.Controller, *mock.MockStandardStorage, *mock.MockAuthorizer, authwrapper.StandardStorage) {
	t.Helper()
	ctrl := gomock.NewController(t)
	store := mock.NewMockStandardStorage(ctrl)
	mauth := mock.NewMockAuthorizer(ctrl)

	subject := mustAuthorizedStorage(t, clusterScopedStandardStorage{store}, gvr, mauth).(authwrapper.StandardStorage)
	return ctrl, store, mauth, subject
}

func allowAuthResponse(mauth *mock.MockAuthorizer) *gomock.Call {
	return mauth.EXPECT().
		Authorize(gomock.Any(), gomock.Any()).
		Return(authorizer.DecisionAllow, "", nil).
		Times(1)
}

func denyAuthResponse(mauth *mock.MockAuthorizer) *gomock.Call {
	return mauth.EXPECT().
		Authorize(gomock.Any(), gomock.Any()).
		Return(authorizer.DecisionDeny, "", nil).
		Times(1)
}

func ctxWithInfo(verb string, name string) context.Context {
	return request.WithUser(
		request.WithRequestInfo(request.NewContext(),
			&request.RequestInfo{
				APIGroup:   gvr.Group,
				APIVersion: gvr.Version,
				Resource:   gvr.Resource,

				Verb: verb,
				Name: name,
			}),
		&user.DefaultInfo{
			Name: "testuser",
		})
}

func mustAuthorizedStorage(t *testing.T, base authwrapper.StorageScoper, rbacID metav1.GroupVersionResource, auth authorizer.Authorizer) authwrapper.Storage {
	t.Helper()

	s, err := authwrapper.NewAuthorizedStorage(base, gvr, auth)
	require.NoError(t, err)
	return s
}

var testUpdateInfoNoUpdate = testUpdateInfo(func(obj runtime.Object) runtime.Object { return obj })

type testUpdateInfo func(obj runtime.Object) runtime.Object

func (testUpdateInfo) Preconditions() *metav1.Preconditions {
	return nil
}
func (ui testUpdateInfo) UpdatedObject(ctx context.Context, oldObj runtime.Object) (newObj runtime.Object, err error) {
	return ui(oldObj), nil
}

type storageWithList struct{ rest.Storage }

func (storageWithList) NewList() runtime.Object {
	return (&testresource.TestResource{}).NewList()
}

func (storageWithList) ConvertToTable(context.Context, runtime.Object, runtime.Object) (*metav1.Table, error) {
	return nil, errors.New("not implemented")
}

type testWatcher struct {
	events chan watch.Event
}

func (w testWatcher) Stop() {}

func (w testWatcher) ResultChan() <-chan watch.Event {
	return w.events
}

type clusterScopedStandardStorage struct {
	rest.StandardStorage
}

func (clusterScopedStandardStorage) NamespaceScoped() bool {
	return false
}

type clusterScopedStorage struct {
	rest.Storage
}

func (clusterScopedStorage) NamespaceScoped() bool {
	return false
}
