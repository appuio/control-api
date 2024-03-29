// Code generated by MockGen. DO NOT EDIT.
// Source: k8s.io/apiserver/pkg/registry/rest (interfaces: StandardStorage,Storage,Responder)
//
// Generated by this command:
//
//	mockgen -destination=./mock/storage.go -package mock k8s.io/apiserver/pkg/registry/rest StandardStorage,Storage,Responder
//
// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
	internalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/apiserver/pkg/registry/rest"
)

// MockStandardStorage is a mock of StandardStorage interface.
type MockStandardStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStandardStorageMockRecorder
}

// MockStandardStorageMockRecorder is the mock recorder for MockStandardStorage.
type MockStandardStorageMockRecorder struct {
	mock *MockStandardStorage
}

// NewMockStandardStorage creates a new mock instance.
func NewMockStandardStorage(ctrl *gomock.Controller) *MockStandardStorage {
	mock := &MockStandardStorage{ctrl: ctrl}
	mock.recorder = &MockStandardStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStandardStorage) EXPECT() *MockStandardStorageMockRecorder {
	return m.recorder
}

// ConvertToTable mocks base method.
func (m *MockStandardStorage) ConvertToTable(arg0 context.Context, arg1, arg2 runtime.Object) (*v1.Table, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConvertToTable", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.Table)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ConvertToTable indicates an expected call of ConvertToTable.
func (mr *MockStandardStorageMockRecorder) ConvertToTable(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConvertToTable", reflect.TypeOf((*MockStandardStorage)(nil).ConvertToTable), arg0, arg1, arg2)
}

// Create mocks base method.
func (m *MockStandardStorage) Create(arg0 context.Context, arg1 runtime.Object, arg2 rest.ValidateObjectFunc, arg3 *v1.CreateOptions) (runtime.Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(runtime.Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockStandardStorageMockRecorder) Create(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockStandardStorage)(nil).Create), arg0, arg1, arg2, arg3)
}

// Delete mocks base method.
func (m *MockStandardStorage) Delete(arg0 context.Context, arg1 string, arg2 rest.ValidateObjectFunc, arg3 *v1.DeleteOptions) (runtime.Object, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(runtime.Object)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Delete indicates an expected call of Delete.
func (mr *MockStandardStorageMockRecorder) Delete(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockStandardStorage)(nil).Delete), arg0, arg1, arg2, arg3)
}

// DeleteCollection mocks base method.
func (m *MockStandardStorage) DeleteCollection(arg0 context.Context, arg1 rest.ValidateObjectFunc, arg2 *v1.DeleteOptions, arg3 *internalversion.ListOptions) (runtime.Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCollection", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(runtime.Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteCollection indicates an expected call of DeleteCollection.
func (mr *MockStandardStorageMockRecorder) DeleteCollection(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCollection", reflect.TypeOf((*MockStandardStorage)(nil).DeleteCollection), arg0, arg1, arg2, arg3)
}

// Destroy mocks base method.
func (m *MockStandardStorage) Destroy() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Destroy")
}

// Destroy indicates an expected call of Destroy.
func (mr *MockStandardStorageMockRecorder) Destroy() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Destroy", reflect.TypeOf((*MockStandardStorage)(nil).Destroy))
}

// Get mocks base method.
func (m *MockStandardStorage) Get(arg0 context.Context, arg1 string, arg2 *v1.GetOptions) (runtime.Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1, arg2)
	ret0, _ := ret[0].(runtime.Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockStandardStorageMockRecorder) Get(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockStandardStorage)(nil).Get), arg0, arg1, arg2)
}

// List mocks base method.
func (m *MockStandardStorage) List(arg0 context.Context, arg1 *internalversion.ListOptions) (runtime.Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", arg0, arg1)
	ret0, _ := ret[0].(runtime.Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockStandardStorageMockRecorder) List(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockStandardStorage)(nil).List), arg0, arg1)
}

// New mocks base method.
func (m *MockStandardStorage) New() runtime.Object {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "New")
	ret0, _ := ret[0].(runtime.Object)
	return ret0
}

// New indicates an expected call of New.
func (mr *MockStandardStorageMockRecorder) New() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "New", reflect.TypeOf((*MockStandardStorage)(nil).New))
}

// NewList mocks base method.
func (m *MockStandardStorage) NewList() runtime.Object {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewList")
	ret0, _ := ret[0].(runtime.Object)
	return ret0
}

// NewList indicates an expected call of NewList.
func (mr *MockStandardStorageMockRecorder) NewList() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewList", reflect.TypeOf((*MockStandardStorage)(nil).NewList))
}

// Update mocks base method.
func (m *MockStandardStorage) Update(arg0 context.Context, arg1 string, arg2 rest.UpdatedObjectInfo, arg3 rest.ValidateObjectFunc, arg4 rest.ValidateObjectUpdateFunc, arg5 bool, arg6 *v1.UpdateOptions) (runtime.Object, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0, arg1, arg2, arg3, arg4, arg5, arg6)
	ret0, _ := ret[0].(runtime.Object)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Update indicates an expected call of Update.
func (mr *MockStandardStorageMockRecorder) Update(arg0, arg1, arg2, arg3, arg4, arg5, arg6 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockStandardStorage)(nil).Update), arg0, arg1, arg2, arg3, arg4, arg5, arg6)
}

// Watch mocks base method.
func (m *MockStandardStorage) Watch(arg0 context.Context, arg1 *internalversion.ListOptions) (watch.Interface, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", arg0, arg1)
	ret0, _ := ret[0].(watch.Interface)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Watch indicates an expected call of Watch.
func (mr *MockStandardStorageMockRecorder) Watch(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockStandardStorage)(nil).Watch), arg0, arg1)
}

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// Destroy mocks base method.
func (m *MockStorage) Destroy() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Destroy")
}

// Destroy indicates an expected call of Destroy.
func (mr *MockStorageMockRecorder) Destroy() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Destroy", reflect.TypeOf((*MockStorage)(nil).Destroy))
}

// New mocks base method.
func (m *MockStorage) New() runtime.Object {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "New")
	ret0, _ := ret[0].(runtime.Object)
	return ret0
}

// New indicates an expected call of New.
func (mr *MockStorageMockRecorder) New() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "New", reflect.TypeOf((*MockStorage)(nil).New))
}

// MockResponder is a mock of Responder interface.
type MockResponder struct {
	ctrl     *gomock.Controller
	recorder *MockResponderMockRecorder
}

// MockResponderMockRecorder is the mock recorder for MockResponder.
type MockResponderMockRecorder struct {
	mock *MockResponder
}

// NewMockResponder creates a new mock instance.
func NewMockResponder(ctrl *gomock.Controller) *MockResponder {
	mock := &MockResponder{ctrl: ctrl}
	mock.recorder = &MockResponderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResponder) EXPECT() *MockResponderMockRecorder {
	return m.recorder
}

// Error mocks base method.
func (m *MockResponder) Error(arg0 error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Error", arg0)
}

// Error indicates an expected call of Error.
func (mr *MockResponderMockRecorder) Error(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockResponder)(nil).Error), arg0)
}

// Object mocks base method.
func (m *MockResponder) Object(arg0 int, arg1 runtime.Object) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Object", arg0, arg1)
}

// Object indicates an expected call of Object.
func (mr *MockResponderMockRecorder) Object(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Object", reflect.TypeOf((*MockResponder)(nil).Object), arg0, arg1)
}
