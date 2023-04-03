package node

import (
	"fmt"

	gomock "github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/infrastructure/rpc"
	"github.com/rootyyang/yangkv/pkg/proto"
)

type NodeMeta struct {
	NodeId      string
	Ip          string
	Port        string
	ClusterName string
	ClusterId   string
}
type ApendEntriesReq struct {
	Term     int
	LeaderId string
}

type ApendEntriesResp struct {
	Term    int
	Success bool
}

type RequstVote struct {
	Term int
}

type RequstVoteResp struct {
	Term        int
	VoteGranted bool
}

type NodeInterface interface {
	Start() error
	Stop() error
	GetNodeMeta() *proto.NodeMeta
	GetRPC() rpc.RPClient
}

type NodeProvider interface {
	GetNode(*proto.NodeMeta) NodeInterface
}
type allNodeName2Provider struct {
	usedNodeProvider      NodeProvider
	nodeName2NodeProvider map[string]NodeProvider
}

var gUseNodeProvider = allNodeName2Provider{nil, make(map[string]NodeProvider)}

func registerNodeProvider(pNodeProviderName string, pNodeProvider NodeProvider) {
	gUseNodeProvider.nodeName2NodeProvider[pNodeProviderName] = pNodeProvider

}

func SetNodeProvider(pProviderName string) error {
	if providerFunc, ok := gUseNodeProvider.nodeName2NodeProvider[pProviderName]; ok {
		gUseNodeProvider.usedNodeProvider = providerFunc
		return nil
	}
	gUseNodeProvider.usedNodeProvider = nil
	return fmt.Errorf("ProviderFunc[%s] Not Found", pProviderName)
}
func GetNodeProvider() (NodeProvider, error) {
	if gUseNodeProvider.usedNodeProvider == nil {
		return nil, fmt.Errorf("Please call SetNodeProvider() first")
	}
	return gUseNodeProvider.usedNodeProvider, nil
}

//以下代码用于Mock测试用
//首先调用Register注册一些用于测试的client节点

func RegisterAndUseMockNode(mockCtl *gomock.Controller) *MockNodeProvider {
	mockNodeProvider := NewMockNodeProvider(mockCtl)
	registerNodeProvider("mockNode", mockNodeProvider)
	SetNodeProvider("mockNode")
	return mockNodeProvider
}
