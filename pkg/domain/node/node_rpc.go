package node

import (
	"context"

	"github.com/rootyyang/yangkv/pkg/infrastructure/rpc"
	"github.com/rootyyang/yangkv/pkg/proto"
)

func init() {
	registerNodeProvider("rpcnode", new(rpcNode))
}

type rpcNode struct {
}

func (*rpcNode) GetNode(pNodeMeta *proto.NodeMeta) NodeInterface {
	return &RPCNode{nodeMeta: pNodeMeta}
}

type RPCNode struct {
	nodeMeta *proto.NodeMeta
	rpc      rpc.RPClient
}

func (rn *RPCNode) Start() error {
	rpcClientProvider, err := rpc.GetRPCClientProvider()
	if err != nil {
		return err
	}

	rn.rpc = rpcClientProvider.GetRPCClient(rn.nodeMeta.Ip, rn.nodeMeta.Port)
	if err := rn.rpc.Start(); err != nil {
		return err
	}
	return nil
}

func (rn *RPCNode) Stop() error {
	if err := rn.rpc.Stop(); err != nil {
		return err
	}
	return nil
}
func (rn *RPCNode) HeartBeat(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error) {
	return rn.rpc.HeartBeat(ctx, req)
}
func (rn *RPCNode) LeaveCluster(ctx context.Context, req *proto.LeaveClusterReq) (*proto.LeaveClusterResp, error) {
	return rn.rpc.LeaveCluster(ctx, req)
}

func (rn *RPCNode) GetNodeMeta() *proto.NodeMeta {
	return rn.nodeMeta
}
