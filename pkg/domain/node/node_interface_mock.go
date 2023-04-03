// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/cluster/node/node_interface.go

// Package node is a generated GoMock package.
package node

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	proto "github.com/rootyyang/yangkv/pkg/proto"
)

// MockNodeInterface is a mock of NodeInterface interface.
type MockNodeInterface struct {
	ctrl     *gomock.Controller
	recorder *MockNodeInterfaceMockRecorder
}

// MockNodeInterfaceMockRecorder is the mock recorder for MockNodeInterface.
type MockNodeInterfaceMockRecorder struct {
	mock *MockNodeInterface
}

// NewMockNodeInterface creates a new mock instance.
func NewMockNodeInterface(ctrl *gomock.Controller) *MockNodeInterface {
	mock := &MockNodeInterface{ctrl: ctrl}
	mock.recorder = &MockNodeInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNodeInterface) EXPECT() *MockNodeInterfaceMockRecorder {
	return m.recorder
}

// GetNodeMeta mocks base method.
func (m *MockNodeInterface) GetNodeMeta() *proto.NodeMeta {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNodeMeta")
	ret0, _ := ret[0].(*proto.NodeMeta)
	return ret0
}

// GetNodeMeta indicates an expected call of GetNodeMeta.
func (mr *MockNodeInterfaceMockRecorder) GetNodeMeta() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodeMeta", reflect.TypeOf((*MockNodeInterface)(nil).GetNodeMeta))
}

// HeartBeat mocks base method.
func (m *MockNodeInterface) HeartBeat(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HeartBeat", ctx, req)
	ret0, _ := ret[0].(*proto.HeartBeatResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HeartBeat indicates an expected call of HeartBeat.
func (mr *MockNodeInterfaceMockRecorder) HeartBeat(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HeartBeat", reflect.TypeOf((*MockNodeInterface)(nil).HeartBeat), ctx, req)
}

// LeaveCluster mocks base method.
func (m *MockNodeInterface) LeaveCluster(ctx context.Context, req *proto.LeaveClusterReq) (*proto.LeaveClusterResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LeaveCluster", ctx, req)
	ret0, _ := ret[0].(*proto.LeaveClusterResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LeaveCluster indicates an expected call of LeaveCluster.
func (mr *MockNodeInterfaceMockRecorder) LeaveCluster(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LeaveCluster", reflect.TypeOf((*MockNodeInterface)(nil).LeaveCluster), ctx, req)
}

// Start mocks base method.
func (m *MockNodeInterface) Start() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockNodeInterfaceMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockNodeInterface)(nil).Start))
}

// Stop mocks base method.
func (m *MockNodeInterface) Stop() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop")
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockNodeInterfaceMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockNodeInterface)(nil).Stop))
}

// MockNodeProvider is a mock of NodeProvider interface.
type MockNodeProvider struct {
	ctrl     *gomock.Controller
	recorder *MockNodeProviderMockRecorder
}

// MockNodeProviderMockRecorder is the mock recorder for MockNodeProvider.
type MockNodeProviderMockRecorder struct {
	mock *MockNodeProvider
}

// NewMockNodeProvider creates a new mock instance.
func NewMockNodeProvider(ctrl *gomock.Controller) *MockNodeProvider {
	mock := &MockNodeProvider{ctrl: ctrl}
	mock.recorder = &MockNodeProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNodeProvider) EXPECT() *MockNodeProviderMockRecorder {
	return m.recorder
}

// GetNode mocks base method.
func (m *MockNodeProvider) GetNode(arg0 *proto.NodeMeta) NodeInterface {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNode", arg0)
	ret0, _ := ret[0].(NodeInterface)
	return ret0
}

// GetNode indicates an expected call of GetNode.
func (mr *MockNodeProviderMockRecorder) GetNode(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNode", reflect.TypeOf((*MockNodeProvider)(nil).GetNode), arg0)
}
