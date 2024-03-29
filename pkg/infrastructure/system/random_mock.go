// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/infrastructure/system/random.go

// Package system is a generated GoMock package.
package system

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockRandom is a mock of Random interface.
type MockRandom struct {
	ctrl     *gomock.Controller
	recorder *MockRandomMockRecorder
}

// MockRandomMockRecorder is the mock recorder for MockRandom.
type MockRandomMockRecorder struct {
	mock *MockRandom
}

// NewMockRandom creates a new mock instance.
func NewMockRandom(ctrl *gomock.Controller) *MockRandom {
	mock := &MockRandom{ctrl: ctrl}
	mock.recorder = &MockRandomMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRandom) EXPECT() *MockRandomMockRecorder {
	return m.recorder
}

// Int mocks base method.
func (m *MockRandom) Int() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Int")
	ret0, _ := ret[0].(int)
	return ret0
}

// Int indicates an expected call of Int.
func (mr *MockRandomMockRecorder) Int() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Int", reflect.TypeOf((*MockRandom)(nil).Int))
}
