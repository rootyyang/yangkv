// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/domain/cluster/node_manager_interface.go

// Package cluster is a generated GoMock package.
package cluster

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	node "github.com/rootyyang/yangkv/pkg/domain/node"
	proto "github.com/rootyyang/yangkv/pkg/proto"
)

// MockNodeManager is a mock of NodeManager interface.
type MockNodeManager struct {
	ctrl     *gomock.Controller
	recorder *MockNodeManagerMockRecorder
}

// MockNodeManagerMockRecorder is the mock recorder for MockNodeManager.
type MockNodeManagerMockRecorder struct {
	mock *MockNodeManager
}

// NewMockNodeManager creates a new mock instance.
func NewMockNodeManager(ctrl *gomock.Controller) *MockNodeManager {
	mock := &MockNodeManager{ctrl: ctrl}
	mock.recorder = &MockNodeManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNodeManager) EXPECT() *MockNodeManagerMockRecorder {
	return m.recorder
}

// GetAllNode mocks base method.
func (m *MockNodeManager) GetAllNode() map[string]node.NodeInterface {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllNode")
	ret0, _ := ret[0].(map[string]node.NodeInterface)
	return ret0
}

// GetAllNode indicates an expected call of GetAllNode.
func (mr *MockNodeManagerMockRecorder) GetAllNode() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllNode", reflect.TypeOf((*MockNodeManager)(nil).GetAllNode))
}

// GetNodeByID mocks base method.
func (m *MockNodeManager) GetNodeByID(pID string) node.NodeInterface {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNodeByID", pID)
	ret0, _ := ret[0].(node.NodeInterface)
	return ret0
}

// GetNodeByID indicates an expected call of GetNodeByID.
func (mr *MockNodeManagerMockRecorder) GetNodeByID(pID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodeByID", reflect.TypeOf((*MockNodeManager)(nil).GetNodeByID), pID)
}

// Init mocks base method.
func (m *MockNodeManager) Init(localNode *proto.NodeMeta) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Init", localNode)
}

// Init indicates an expected call of Init.
func (mr *MockNodeManagerMockRecorder) Init(localNode interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockNodeManager)(nil).Init), localNode)
}

// JoinCluster mocks base method.
func (m *MockNodeManager) JoinCluster(arg0 *proto.NodeMeta) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "JoinCluster", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// JoinCluster indicates an expected call of JoinCluster.
func (mr *MockNodeManagerMockRecorder) JoinCluster(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "JoinCluster", reflect.TypeOf((*MockNodeManager)(nil).JoinCluster), arg0)
}

// LeftCluster mocks base method.
func (m *MockNodeManager) LeftCluster(arg0 *proto.NodeMeta) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LeftCluster", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// LeftCluster indicates an expected call of LeftCluster.
func (mr *MockNodeManagerMockRecorder) LeftCluster(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LeftCluster", reflect.TypeOf((*MockNodeManager)(nil).LeftCluster), arg0)
}

// RegisterAfterLeftFunc mocks base method.
func (m *MockNodeManager) RegisterAfterLeftFunc() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterAfterLeftFunc")
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterAfterLeftFunc indicates an expected call of RegisterAfterLeftFunc.
func (mr *MockNodeManagerMockRecorder) RegisterAfterLeftFunc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterAfterLeftFunc", reflect.TypeOf((*MockNodeManager)(nil).RegisterAfterLeftFunc))
}

// RegisterBeforeJoinFunc mocks base method.
func (m *MockNodeManager) RegisterBeforeJoinFunc() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterBeforeJoinFunc")
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterBeforeJoinFunc indicates an expected call of RegisterBeforeJoinFunc.
func (mr *MockNodeManagerMockRecorder) RegisterBeforeJoinFunc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterBeforeJoinFunc", reflect.TypeOf((*MockNodeManager)(nil).RegisterBeforeJoinFunc))
}

// Start mocks base method.
func (m *MockNodeManager) Start() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockNodeManagerMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockNodeManager)(nil).Start))
}

// Stop mocks base method.
func (m *MockNodeManager) Stop() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop")
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockNodeManagerMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockNodeManager)(nil).Stop))
}

// UnregisterAfterLeftFunc mocks base method.
func (m *MockNodeManager) UnregisterAfterLeftFunc() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnregisterAfterLeftFunc")
	ret0, _ := ret[0].(error)
	return ret0
}

// UnregisterAfterLeftFunc indicates an expected call of UnregisterAfterLeftFunc.
func (mr *MockNodeManagerMockRecorder) UnregisterAfterLeftFunc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnregisterAfterLeftFunc", reflect.TypeOf((*MockNodeManager)(nil).UnregisterAfterLeftFunc))
}

// UnregisterBeforeJoinFunc mocks base method.
func (m *MockNodeManager) UnregisterBeforeJoinFunc() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnregisterBeforeJoinFunc")
	ret0, _ := ret[0].(error)
	return ret0
}

// UnregisterBeforeJoinFunc indicates an expected call of UnregisterBeforeJoinFunc.
func (mr *MockNodeManagerMockRecorder) UnregisterBeforeJoinFunc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnregisterBeforeJoinFunc", reflect.TypeOf((*MockNodeManager)(nil).UnregisterBeforeJoinFunc))
}
