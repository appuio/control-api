// Code generated by MockGen. DO NOT EDIT.
// Source: authorizer.go

// Package mock_organization is a generated GoMock package.
package mock_organization

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	authorizer "k8s.io/apiserver/pkg/authorization/authorizer"
)

// MockContextAuthorizer is a mock of ContextAuthorizer interface.
type MockContextAuthorizer struct {
	ctrl     *gomock.Controller
	recorder *MockContextAuthorizerMockRecorder
}

// MockContextAuthorizerMockRecorder is the mock recorder for MockContextAuthorizer.
type MockContextAuthorizerMockRecorder struct {
	mock *MockContextAuthorizer
}

// NewMockContextAuthorizer creates a new mock instance.
func NewMockContextAuthorizer(ctrl *gomock.Controller) *MockContextAuthorizer {
	mock := &MockContextAuthorizer{ctrl: ctrl}
	mock.recorder = &MockContextAuthorizerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockContextAuthorizer) EXPECT() *MockContextAuthorizerMockRecorder {
	return m.recorder
}

// Authorize mocks base method.
func (m *MockContextAuthorizer) Authorize(ctx context.Context, a authorizer.Attributes) (authorizer.Decision, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Authorize", ctx, a)
	ret0, _ := ret[0].(authorizer.Decision)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Authorize indicates an expected call of Authorize.
func (mr *MockContextAuthorizerMockRecorder) Authorize(ctx, a interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Authorize", reflect.TypeOf((*MockContextAuthorizer)(nil).Authorize), ctx, a)
}

// AuthorizeContext mocks base method.
func (m *MockContextAuthorizer) AuthorizeContext(ctx context.Context) (authorizer.Decision, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AuthorizeContext", ctx)
	ret0, _ := ret[0].(authorizer.Decision)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// AuthorizeContext indicates an expected call of AuthorizeContext.
func (mr *MockContextAuthorizerMockRecorder) AuthorizeContext(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AuthorizeContext", reflect.TypeOf((*MockContextAuthorizer)(nil).AuthorizeContext), ctx)
}
