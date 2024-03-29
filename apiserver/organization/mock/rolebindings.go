// Code generated by MockGen. DO NOT EDIT.
// Source: rolebindings.go
//
// Generated by this command:
//
//	mockgen -source=rolebindings.go -destination=./mock/rolebindings.go
//
// Package mock_organization is a generated GoMock package.
package mock_organization

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockroleBindingCreator is a mock of roleBindingCreator interface.
type MockroleBindingCreator struct {
	ctrl     *gomock.Controller
	recorder *MockroleBindingCreatorMockRecorder
}

// MockroleBindingCreatorMockRecorder is the mock recorder for MockroleBindingCreator.
type MockroleBindingCreatorMockRecorder struct {
	mock *MockroleBindingCreator
}

// NewMockroleBindingCreator creates a new mock instance.
func NewMockroleBindingCreator(ctrl *gomock.Controller) *MockroleBindingCreator {
	mock := &MockroleBindingCreator{ctrl: ctrl}
	mock.recorder = &MockroleBindingCreatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockroleBindingCreator) EXPECT() *MockroleBindingCreatorMockRecorder {
	return m.recorder
}

// CreateRoleBindings mocks base method.
func (m *MockroleBindingCreator) CreateRoleBindings(ctx context.Context, namespace, username string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateRoleBindings", ctx, namespace, username)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateRoleBindings indicates an expected call of CreateRoleBindings.
func (mr *MockroleBindingCreatorMockRecorder) CreateRoleBindings(ctx, namespace, username any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateRoleBindings", reflect.TypeOf((*MockroleBindingCreator)(nil).CreateRoleBindings), ctx, namespace, username)
}
