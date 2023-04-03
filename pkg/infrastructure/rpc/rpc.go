package rpc

import (
	"context"
	"fmt"

	gomock "github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/proto"
)

//RPCClient和RPCServer应该是线程安全的

type RPCClientProvider interface {
	GetRPCClient(remoteNodeIP string, remoteNodePort string) RPClient
}

type rpcWrapper struct {
	rpcProvider RPCClientProvider
	rpcServer   RPCServer
}
type allRPCAndUsedRPC struct {
	usedRPC             rpcWrapper
	rpcName2RPCProvider map[string]rpcWrapper
}

var gUseRPC = allRPCAndUsedRPC{rpcWrapper{nil, nil}, make(map[string]rpcWrapper)}

//rpc实现，尽量放到rpc目录下，所以注册函数首字母小写
func registerRPC(pRPCName string, pRPCClientProvider RPCClientProvider, pRPCServer RPCServer) {
	gUseRPC.rpcName2RPCProvider[pRPCName] = rpcWrapper{pRPCClientProvider, pRPCServer}
}

func SetUseRPC(RPCName string) error {
	if value, ok := gUseRPC.rpcName2RPCProvider[RPCName]; ok {
		gUseRPC.usedRPC = value
		return nil
	}
	gUseRPC.usedRPC = rpcWrapper{nil, nil}
	return fmt.Errorf("RPC[%v] not found", RPCName)
}

func GetRPCClientProvider() (RPCClientProvider, error) {
	if gUseRPC.usedRPC.rpcProvider == nil {
		return nil, fmt.Errorf("Please call SetUseRPC() first")
	}
	return gUseRPC.usedRPC.rpcProvider, nil
}

func GetRPCServer() (RPCServer, error) {
	if gUseRPC.usedRPC.rpcServer == nil {
		return nil, fmt.Errorf("Please call SetUseRPC() first")
	}

	return gUseRPC.usedRPC.rpcServer, nil
}

type RPClient interface {
	Start() error
	Stop() error
	HeartBeat(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error)
	LeaveCluster(ctx context.Context, req *proto.LeaveClusterReq) (*proto.LeaveClusterResp, error)

	AppendEntires(ctx context.Context, req *AppendEntiresReq) (*AppendEtriesResp, error)
	RequestVote(ctx context.Context, req *RequestVoteReq) (*RequestVoteResp, error)
}

type RPCSeverHandleInterface interface {
	Start() error
	Stop() error
	HeartBeat(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error)
	LeaveCluster(ctx context.Context, req *proto.LeaveClusterReq) (*proto.LeaveClusterResp, error)
}

type RPCServer interface {
	RegisterRaftHandle(pRaftHandler RaftServiceHandler) error
	Start(pPort string, pRPCHandleFunc RPCSeverHandleInterface) error

	Stop() error
}

//以下代码用于Mock测试用
//首先调用Register注册一些用于测试的client节点
func RegisterAndUseMockRPC(mockCtl *gomock.Controller) (*MockRPCClientProvider, *MockRPCServerInterface) {
	mockRPCServerInterface := NewMockRPCServerInterface(mockCtl)
	mockRPCProvider := NewMockRPCClientProvider(mockCtl)
	registerRPC("mockRPC", mockRPCProvider, mockRPCServerInterface)
	SetUseRPC("mockRPC")
	return mockRPCProvider, mockRPCServerInterface
}
