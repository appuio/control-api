// Code generated by MockGen. DO NOT EDIT.
// Source: controllers/saleorder/saleorder_storage.go
//
// Generated by this command:
//
//	mockgen -source=controllers/saleorder/saleorder_storage.go
//
// Package mock_saleorder is a generated GoMock package.
package mock_saleorder

import (
	reflect "reflect"

	v1 "github.com/appuio/control-api/apis/organization/v1"
	odoo "github.com/appuio/go-odoo"
	gomock "go.uber.org/mock/gomock"
)

// MockSaleOrderStorage is a mock of SaleOrderStorage interface.
type MockSaleOrderStorage struct {
	ctrl     *gomock.Controller
	recorder *MockSaleOrderStorageMockRecorder
}

// MockSaleOrderStorageMockRecorder is the mock recorder for MockSaleOrderStorage.
type MockSaleOrderStorageMockRecorder struct {
	mock *MockSaleOrderStorage
}

// NewMockSaleOrderStorage creates a new mock instance.
func NewMockSaleOrderStorage(ctrl *gomock.Controller) *MockSaleOrderStorage {
	mock := &MockSaleOrderStorage{ctrl: ctrl}
	mock.recorder = &MockSaleOrderStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSaleOrderStorage) EXPECT() *MockSaleOrderStorageMockRecorder {
	return m.recorder
}

// CreateSaleOrder mocks base method.
func (m *MockSaleOrderStorage) CreateSaleOrder(arg0 v1.Organization) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSaleOrder", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateSaleOrder indicates an expected call of CreateSaleOrder.
func (mr *MockSaleOrderStorageMockRecorder) CreateSaleOrder(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSaleOrder", reflect.TypeOf((*MockSaleOrderStorage)(nil).CreateSaleOrder), arg0)
}

// GetSaleOrderName mocks base method.
func (m *MockSaleOrderStorage) GetSaleOrderName(arg0 v1.Organization) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSaleOrderName", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSaleOrderName indicates an expected call of GetSaleOrderName.
func (mr *MockSaleOrderStorageMockRecorder) GetSaleOrderName(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSaleOrderName", reflect.TypeOf((*MockSaleOrderStorage)(nil).GetSaleOrderName), arg0)
}

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

// CreateSaleOrder mocks base method.
func (m *MockOdoo16Client) CreateSaleOrder(arg0 *odoo.SaleOrder) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSaleOrder", arg0)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateSaleOrder indicates an expected call of CreateSaleOrder.
func (mr *MockOdoo16ClientMockRecorder) CreateSaleOrder(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSaleOrder", reflect.TypeOf((*MockOdoo16Client)(nil).CreateSaleOrder), arg0)
}

// Read mocks base method.
func (m *MockOdoo16Client) Read(arg0 string, arg1 []int64, arg2 *odoo.Options, arg3 any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// Read indicates an expected call of Read.
func (mr *MockOdoo16ClientMockRecorder) Read(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockOdoo16Client)(nil).Read), arg0, arg1, arg2, arg3)
}
