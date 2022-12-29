package authwrapper

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/registry/rest"
)

//go:generate go run github.com/golang/mock/mockgen -destination=./mock/$GOFILE -package mock k8s.io/apiserver/pkg/registry/rest StandardStorage,Storage

var _ rest.StandardStorage = &authorizedStorageWithLister{}

type Storage interface {
	rest.Storage

	// Storage returns the underlying storage
	Storage() rest.Storage
}

type StandardStorage interface {
	rest.StandardStorage

	// Storage returns the underlying storage
	Storage() rest.Storage
}

type StorageScoper interface {
	rest.Storage
	rest.Scoper
}

// authorizedStorage is a wrapper around a rest.Storage
// authorizing all requests and implementing the rest.Storage interface
type authorizedStorage struct {
	storage    rest.Storage
	authorizer Authorizer
}

// authorizedStorageWithLister is a wrapper around a rest.StandardStorage that
type authorizedStorageWithLister struct {
	*authorizedStorage
}

// NewList implements rest.Lister/rest.StandardStorage
func (s *authorizedStorageWithLister) NewList() runtime.Object {
	return s.authorizedStorage.storage.(rest.Lister).NewList()
}

// ConvertToTable implements rest.Lister/rest.StandardStorage
func (s *authorizedStorageWithLister) ConvertToTable(ctx context.Context, obj runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return s.authorizedStorage.storage.(rest.Lister).ConvertToTable(ctx, obj, tableOptions)
}

// NewAuthorizedStorage returns a new wrapper around the given storage
// authorizing all requests based on rbacNamespace and implementing the rest.Storage or rest.StandardStorage interface.
// It allows filtering list and watch results based on the user's RBAC permissions.
// If the storage implements rest.StandardStorage, the returned storage will implement rest.StandardStorage.
// If the storage implements rest.Storage, the returned storage will implement rest.Storage.
// Only cluster-scoped resources currently are supported. Panics if the storage is namespace-scoped.
func NewAuthorizedStorage(storage StorageScoper, rbacNamespace metav1.GroupVersionResource, auth authorizer.Authorizer) (Storage, error) {
	if storage.NamespaceScoped() {
		return nil, errors.New("namespace-scoped resources are not supported")
	}
	s := &authorizedStorage{
		storage:    storage,
		authorizer: NewAuthorizer(rbacNamespace, auth),
	}
	if _, ok := storage.(rest.Lister); ok {
		return &authorizedStorageWithLister{s}, nil
	}
	return s, nil
}

// New implements rest.Storage
func (s *authorizedStorage) New() runtime.Object {
	return s.storage.New()
}

// Destroy implements rest.Storage
func (s *authorizedStorage) Destroy() {
	s.storage.Destroy()
}

// Storage returns the underlying storage
func (s *authorizedStorage) Storage() rest.Storage {
	return s.storage
}

// NamespaceScoped implements rest.Scoper
// Only cluster-scoped resources are supported.
func (s *authorizedStorage) NamespaceScoped() bool {
	return false
}

func (s *authorizedStorage) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, opts *metav1.CreateOptions) (runtime.Object, error) {
	err := s.authorizer.AuthorizeContext(ctx)
	if err != nil {
		return nil, err
	}

	if stor, ok := s.storage.(rest.Creater); ok {
		return stor.Create(ctx, obj, createValidation, opts)
	}

	return nil, newMethodNotSupported("create")
}

func (s *authorizedStorage) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	err := s.authorizer.AuthorizeContext(ctx)
	if err != nil {
		return nil, false, err
	}

	if stor, ok := s.storage.(rest.GracefulDeleter); ok {
		return stor.Delete(ctx, name, deleteValidation, options)
	}

	return nil, false, newMethodNotSupported("delete")
}

func (s *authorizedStorage) DeleteCollection(ctx context.Context, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions, listOptions *metainternalversion.ListOptions) (runtime.Object, error) {
	// TODO(swi): Implement DeleteCollection
	return nil, newMethodNotSupported("deletecollection")
}

func (s *authorizedStorage) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	err := s.authorizer.AuthorizeContext(ctx)
	if err != nil {
		return nil, err
	}

	if stor, ok := s.storage.(rest.Getter); ok {
		return stor.Get(ctx, name, options)
	}

	return nil, newMethodNotSupported("get")
}

func (s *authorizedStorageWithLister) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	err := s.authorizer.AuthorizeContext(ctx)
	if err != nil {
		return nil, err
	}

	stor := s.storage.(rest.Lister)
	obj, err := stor.List(ctx, options)
	if err != nil {
		return nil, err
	}

	l, err := apimeta.ExtractList(obj)
	if err != nil {
		return nil, err
	}

	ac := apimeta.NewAccessor()
	filtered := make([]runtime.Object, 0, len(l))
	for _, itm := range l {
		name, err := ac.Name(itm)
		if err != nil {
			return nil, err
		}

		if err := s.authorizer.AuthorizeGet(ctx, name); err != nil {
			continue
		}

		filtered = append(filtered, itm)
	}

	fl := stor.NewList()
	apimeta.SetList(fl, filtered)
	return fl, nil
}

func (s *authorizedStorage) Watch(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	err := s.authorizer.AuthorizeContext(ctx)
	if err != nil {
		return nil, err
	}

	stor, ok := s.storage.(rest.Watcher)
	if !ok {
		return nil, newMethodNotSupported("watch")
	}

	watcher, err := stor.Watch(ctx, options)
	if err != nil {
		return nil, err
	}

	ac := apimeta.NewAccessor()
	return watch.Filter(watcher, func(in watch.Event) (out watch.Event, keep bool) {
		if in.Type == watch.Error || in.Type == watch.Bookmark || in.Object == nil {
			return in, true
		}

		// Drop if name can't be accessed
		name, err := ac.Name(in.Object)
		if err != nil {
			return in, false
		}
		if err := s.authorizer.AuthorizeGet(ctx, name); err != nil {
			return in, false
		}

		return in, true
	}), nil
}

func (s *authorizedStorage) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo,
	createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	err := s.authorizer.AuthorizeContext(ctx)
	if err != nil {
		return nil, false, err
	}

	if stor, ok := s.storage.(rest.Updater); ok {
		return stor.Update(ctx, name, objInfo, createValidation, updateValidation, forceAllowCreate, options)
	}

	return nil, false, newMethodNotSupported("update")
}

func newMethodNotSupported(action string) *apierrors.StatusError {
	return &apierrors.StatusError{
		ErrStatus: metav1.Status{
			Status:  metav1.StatusFailure,
			Code:    http.StatusMethodNotAllowed,
			Reason:  metav1.StatusReasonMethodNotAllowed,
			Message: fmt.Sprintf("%s is not supported", action),
		}}
}
