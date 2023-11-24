// Code generated by MockGen. DO NOT EDIT.
// Source: members.go
//
// Generated by this command:
//
//	mockgen -source=members.go -destination=./mock/members.go
//
// Package mock_organization is a generated GoMock package.
package mock_organization

import (
	context "context"
	reflect "reflect"

	v1 "github.com/appuio/control-api/apis/v1"
	gomock "go.uber.org/mock/gomock"
)

// MockmemberProvider is a mock of memberProvider interface.
type MockmemberProvider struct {
	ctrl     *gomock.Controller
	recorder *MockmemberProviderMockRecorder
}

// MockmemberProviderMockRecorder is the mock recorder for MockmemberProvider.
type MockmemberProviderMockRecorder struct {
	mock *MockmemberProvider
}

// NewMockmemberProvider creates a new mock instance.
func NewMockmemberProvider(ctrl *gomock.Controller) *MockmemberProvider {
	mock := &MockmemberProvider{ctrl: ctrl}
	mock.recorder = &MockmemberProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockmemberProvider) EXPECT() *MockmemberProviderMockRecorder {
	return m.recorder
}

// CreateMembers mocks base method.
func (m *MockmemberProvider) CreateMembers(ctx context.Context, members *v1.OrganizationMembers) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMembers", ctx, members)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateMembers indicates an expected call of CreateMembers.
func (mr *MockmemberProviderMockRecorder) CreateMembers(ctx, members any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMembers", reflect.TypeOf((*MockmemberProvider)(nil).CreateMembers), ctx, members)
}
