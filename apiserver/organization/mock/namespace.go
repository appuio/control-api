// Code generated by MockGen. DO NOT EDIT.
// Source: namespace.go

// Package mock_organization is a generated GoMock package.
package mock_organization

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
	internalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	v10 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch "k8s.io/apimachinery/pkg/watch"
)

// MocknamespaceProvider is a mock of namespaceProvider interface.
type MocknamespaceProvider struct {
	ctrl     *gomock.Controller
	recorder *MocknamespaceProviderMockRecorder
}

// MocknamespaceProviderMockRecorder is the mock recorder for MocknamespaceProvider.
type MocknamespaceProviderMockRecorder struct {
	mock *MocknamespaceProvider
}

// NewMocknamespaceProvider creates a new mock instance.
func NewMocknamespaceProvider(ctrl *gomock.Controller) *MocknamespaceProvider {
	mock := &MocknamespaceProvider{ctrl: ctrl}
	mock.recorder = &MocknamespaceProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MocknamespaceProvider) EXPECT() *MocknamespaceProviderMockRecorder {
	return m.recorder
}

// CreateNamespace mocks base method.
func (m *MocknamespaceProvider) CreateNamespace(ctx context.Context, ns *v1.Namespace, options *v10.CreateOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateNamespace", ctx, ns, options)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateNamespace indicates an expected call of CreateNamespace.
func (mr *MocknamespaceProviderMockRecorder) CreateNamespace(ctx, ns, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateNamespace", reflect.TypeOf((*MocknamespaceProvider)(nil).CreateNamespace), ctx, ns, options)
}

// DeleteNamespace mocks base method.
func (m *MocknamespaceProvider) DeleteNamespace(ctx context.Context, name string, options *v10.DeleteOptions) (*v1.Namespace, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteNamespace", ctx, name, options)
	ret0, _ := ret[0].(*v1.Namespace)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteNamespace indicates an expected call of DeleteNamespace.
func (mr *MocknamespaceProviderMockRecorder) DeleteNamespace(ctx, name, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteNamespace", reflect.TypeOf((*MocknamespaceProvider)(nil).DeleteNamespace), ctx, name, options)
}

// GetNamespace mocks base method.
func (m *MocknamespaceProvider) GetNamespace(ctx context.Context, name string, options *v10.GetOptions) (*v1.Namespace, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNamespace", ctx, name, options)
	ret0, _ := ret[0].(*v1.Namespace)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNamespace indicates an expected call of GetNamespace.
func (mr *MocknamespaceProviderMockRecorder) GetNamespace(ctx, name, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNamespace", reflect.TypeOf((*MocknamespaceProvider)(nil).GetNamespace), ctx, name, options)
}

// ListNamespaces mocks base method.
func (m *MocknamespaceProvider) ListNamespaces(ctx context.Context, options *internalversion.ListOptions) (*v1.NamespaceList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListNamespaces", ctx, options)
	ret0, _ := ret[0].(*v1.NamespaceList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListNamespaces indicates an expected call of ListNamespaces.
func (mr *MocknamespaceProviderMockRecorder) ListNamespaces(ctx, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListNamespaces", reflect.TypeOf((*MocknamespaceProvider)(nil).ListNamespaces), ctx, options)
}

// UpdateNamespace mocks base method.
func (m *MocknamespaceProvider) UpdateNamespace(ctx context.Context, ns *v1.Namespace, options *v10.UpdateOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNamespace", ctx, ns, options)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateNamespace indicates an expected call of UpdateNamespace.
func (mr *MocknamespaceProviderMockRecorder) UpdateNamespace(ctx, ns, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNamespace", reflect.TypeOf((*MocknamespaceProvider)(nil).UpdateNamespace), ctx, ns, options)
}

// WatchNamespaces mocks base method.
func (m *MocknamespaceProvider) WatchNamespaces(ctx context.Context, options *internalversion.ListOptions) (watch.Interface, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WatchNamespaces", ctx, options)
	ret0, _ := ret[0].(watch.Interface)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WatchNamespaces indicates an expected call of WatchNamespaces.
func (mr *MocknamespaceProviderMockRecorder) WatchNamespaces(ctx, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WatchNamespaces", reflect.TypeOf((*MocknamespaceProvider)(nil).WatchNamespaces), ctx, options)
}