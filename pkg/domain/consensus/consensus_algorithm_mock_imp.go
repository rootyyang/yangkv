// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/domain/consensus/consensus_algorithm.go

// Package consensus is a generated GoMock package.
package consensus

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockCommand is a mock of Command interface.
type MockCommand struct {
	ctrl     *gomock.Controller
	recorder *MockCommandMockRecorder
}

// MockCommandMockRecorder is the mock recorder for MockCommand.
type MockCommandMockRecorder struct {
	mock *MockCommand
}

// NewMockCommand creates a new mock instance.
func NewMockCommand(ctrl *gomock.Controller) *MockCommand {
	mock := &MockCommand{ctrl: ctrl}
	mock.recorder = &MockCommandMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCommand) EXPECT() *MockCommandMockRecorder {
	return m.recorder
}

// ParseFrom mocks base method.
func (m *MockCommand) ParseFrom(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParseFrom", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ParseFrom indicates an expected call of ParseFrom.
func (mr *MockCommandMockRecorder) ParseFrom(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseFrom", reflect.TypeOf((*MockCommand)(nil).ParseFrom), arg0)
}

// ToString mocks base method.
func (m *MockCommand) ToString() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ToString")
	ret0, _ := ret[0].(string)
	return ret0
}

// ToString indicates an expected call of ToString.
func (mr *MockCommandMockRecorder) ToString() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ToString", reflect.TypeOf((*MockCommand)(nil).ToString))
}

// MockStateMachine is a mock of StateMachine interface.
type MockStateMachine struct {
	ctrl     *gomock.Controller
	recorder *MockStateMachineMockRecorder
}

// MockStateMachineMockRecorder is the mock recorder for MockStateMachine.
type MockStateMachineMockRecorder struct {
	mock *MockStateMachine
}

// NewMockStateMachine creates a new mock instance.
func NewMockStateMachine(ctrl *gomock.Controller) *MockStateMachine {
	mock := &MockStateMachine{ctrl: ctrl}
	mock.recorder = &MockStateMachineMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStateMachine) EXPECT() *MockStateMachineMockRecorder {
	return m.recorder
}

// ApplyLog mocks base method.
func (m *MockStateMachine) ApplyLog(pLog Command) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyLog", pLog)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyLog indicates an expected call of ApplyLog.
func (mr *MockStateMachineMockRecorder) ApplyLog(pLog interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyLog", reflect.TypeOf((*MockStateMachine)(nil).ApplyLog), pLog)
}