package node

import (
	"github.com/rootyyang/yangkv/pkg/infrastructure/rpc"
)

type NodeMeta struct {
	NodeId      string
	Ip          string
	Port        string
	ClusterName string
	ClusterId   string
}

type Node interface {
	Start() error
	Stop() error
	GetNodeMeta() *NodeMeta
	GetRPC() rpc.RPClient
}

type NodeProvider interface {
	GetNode(NodeMeta) Node
}

var gNodeProvider NodeProvider = &nodeImpProvider{}

func GetNodeProvider() NodeProvider {
	return gNodeProvider
}

type nodeImpProvider struct {
}

func (*nodeImpProvider) GetNode(pNodeMeta NodeMeta) Node {
	return &nodeImp{nodeMeta: pNodeMeta}
}

type nodeImp struct {
	nodeMeta NodeMeta
	rpc      rpc.RPClient
}

func (rn *nodeImp) Start() error {
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

func (rn *nodeImp) Stop() error {
	if err := rn.rpc.Stop(); err != nil {
		return err
	}
	return nil
}

func (rn *nodeImp) GetNodeMeta() *NodeMeta {
	return &rn.nodeMeta
}

func (rn *nodeImp) GetRPC() rpc.RPClient {
	return rn.rpc
}

//以下代码用于Mock测试用
//首先调用Register注册一些用于测试的client节点
