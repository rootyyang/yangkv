package rpc

import (
	"context"
	"fmt"
)

//RPCClient和RPCServer应该是线程安全的
//TODO Go网络库和Go RPC多了解一下

type RPCClientProvider interface {
	GetRPCClient(remoteNodeIP string, remoteNodePort string) RPClient
}
type RPCServerProvider interface {
	GetRPCServer(localNodePort string) RPCServer
}

type RPClient interface {
	Start() error
	Stop() error
	AppendEntires(ctx context.Context, req *AppendEntiresReq) (*AppendEntiresResp, error)
	RequestVote(ctx context.Context, req *RequestVoteReq) (*RequestVoteResp, error)
}

type RPCServer interface {
	Start() error
	Stop() error
	RegisterRaftHandle(pRaftHandler RaftServiceHandler) error
}

type rpcWrapper struct {
	rpcClientProvider RPCClientProvider
	rpcServerProvider RPCServerProvider
}
type allRPCAndUsedRPC struct {
	usedRPC             rpcWrapper
	rpcName2RPCProvider map[string]rpcWrapper
}

var gUseRPC = allRPCAndUsedRPC{rpcWrapper{nil, nil}, make(map[string]rpcWrapper)}

//rpc实现，尽量放到rpc目录下，所以注册函数首字母小写
func registerRPC(pRPCName string, pRPCClientProvider RPCClientProvider, pRPCServerProvider RPCServerProvider) {
	gUseRPC.rpcName2RPCProvider[pRPCName] = rpcWrapper{pRPCClientProvider, pRPCServerProvider}
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
	if gUseRPC.usedRPC.rpcClientProvider == nil {
		return nil, fmt.Errorf("Please call SetUseRPC() first")
	}
	return gUseRPC.usedRPC.rpcClientProvider, nil
}

func GetRPCServerProvider() (RPCServerProvider, error) {
	if gUseRPC.usedRPC.rpcServerProvider == nil {
		return nil, fmt.Errorf("Please call SetUseRPC() first")
	}

	return gUseRPC.usedRPC.rpcServerProvider, nil
}

//以下代码用于Mock测试用
//首先调用Register注册一些用于测试的client节点
