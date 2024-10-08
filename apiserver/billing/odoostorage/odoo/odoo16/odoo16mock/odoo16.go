// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo16 (interfaces: Odoo16Client)
//
// Generated by this command:
//
//	mockgen -destination=./odoo16mock/odoo16.go -package odoo16mock . Odoo16Client
//
// Package odoo16mock is a generated GoMock package.
package odoo16mock

import (
	reflect "reflect"

	odoo "github.com/appuio/go-odoo"
	gomock "go.uber.org/mock/gomock"
)

// MockOdoo16Client is a mock of Odoo16Client interface.
type MockOdoo16Client struct {
	ctrl     *gomock.Controller
	recorder *MockOdoo16ClientMockRecorder
}

// MockOdoo16ClientMockRecorder is the mock recorder for MockOdoo16Client.
type MockOdoo16ClientMockRecorder struct {
	mock *MockOdoo16Client
}

// NewMockOdoo16Client creates a new mock instance.
func NewMockOdoo16Client(ctrl *gomock.Controller) *MockOdoo16Client {
	mock := &MockOdoo16Client{ctrl: ctrl}
	mock.recorder = &MockOdoo16ClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOdoo16Client) EXPECT() *MockOdoo16ClientMockRecorder {
	return m.recorder
}

// CreateResPartner mocks base method.
func (m *MockOdoo16Client) CreateResPartner(arg0 *odoo.ResPartner) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateResPartner", arg0)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateResPartner indicates an expected call of CreateResPartner.
func (mr *MockOdoo16ClientMockRecorder) CreateResPartner(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateResPartner", reflect.TypeOf((*MockOdoo16Client)(nil).CreateResPartner), arg0)
}

// DeleteResPartners mocks base method.
func (m *MockOdoo16Client) DeleteResPartners(arg0 []int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteResPartners", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteResPartners indicates an expected call of DeleteResPartners.
func (mr *MockOdoo16ClientMockRecorder) DeleteResPartners(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteResPartners", reflect.TypeOf((*MockOdoo16Client)(nil).DeleteResPartners), arg0)
}

// FindResPartners mocks base method.
func (m *MockOdoo16Client) FindResPartners(arg0 *odoo.Criteria, arg1 *odoo.Options) (*odoo.ResPartners, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindResPartners", arg0, arg1)
	ret0, _ := ret[0].(*odoo.ResPartners)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindResPartners indicates an expected call of FindResPartners.
func (mr *MockOdoo16ClientMockRecorder) FindResPartners(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindResPartners", reflect.TypeOf((*MockOdoo16Client)(nil).FindResPartners), arg0, arg1)
}

// FullInitialization mocks base method.
func (m *MockOdoo16Client) FullInitialization() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FullInitialization")
	ret0, _ := ret[0].(error)
	return ret0
}

// FullInitialization indicates an expected call of FullInitialization.
func (mr *MockOdoo16ClientMockRecorder) FullInitialization() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FullInitialization", reflect.TypeOf((*MockOdoo16Client)(nil).FullInitialization))
}

// Update mocks base method.
func (m *MockOdoo16Client) Update(arg0 string, arg1 []int64, arg2 any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockOdoo16ClientMockRecorder) Update(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockOdoo16Client)(nil).Update), arg0, arg1, arg2)
}

// UpdateResPartner mocks base method.
func (m *MockOdoo16Client) UpdateResPartner(arg0 *odoo.ResPartner) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateResPartner", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateResPartner indicates an expected call of UpdateResPartner.
func (mr *MockOdoo16ClientMockRecorder) UpdateResPartner(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateResPartner", reflect.TypeOf((*MockOdoo16Client)(nil).UpdateResPartner), arg0)
}
