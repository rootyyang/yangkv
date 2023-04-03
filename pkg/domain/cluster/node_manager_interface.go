package cluster

import (
	"fmt"

	gomock "github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/domain/node"
	"github.com/rootyyang/yangkv/pkg/proto"
)

//优化，将业务代码与底层实现解耦，比如，将心跳，NodeMeta与底层实现解耦，不要使用pb

type NodeManager interface {
	Init(localNode *proto.NodeMeta)
	Start() error
	Stop() error
	RegisterBeforeJoinFunc() error
	UnregisterBeforeJoinFunc() error
	RegisterAfterLeftFunc() error
	UnregisterAfterLeftFunc() error
	JoinCluster(*proto.NodeMeta) error
	LeftCluster(*proto.NodeMeta) error
	GetAllNode() map[string]node.NodeInterface
	GetNodeByID(pID string) node.NodeInterface
}

type nodeManagerName2Provider struct {
	usedNodeManager             NodeManager
	nodeManagerName2NodeManager map[string]NodeManager
}

var gUseNodeManager = nodeManagerName2Provider{nil, make(map[string]NodeManager)}

func registerNodeManager(pNodeManagerName string, pNodeManager NodeManager) {
	gUseNodeManager.nodeManagerName2NodeManager[pNodeManagerName] = pNodeManager

}

func SetNodeManager(pNodeManagerName string) error {
	if providerFunc, ok := gUseNodeManager.nodeManagerName2NodeManager[pNodeManagerName]; ok {
		gUseNodeManager.usedNodeManager = providerFunc
		return nil
	}
	gUseNodeManager.usedNodeManager = nil
	return fmt.Errorf("NodeManager[%v] not found", pNodeManagerName)
}
func GetNodeManager() (NodeManager, error) {
	if gUseNodeManager.usedNodeManager == nil {
		return nil, fmt.Errorf("Please call SetNodeManager() first")
	}
	return gUseNodeManager.usedNodeManager, nil
}

func RegisterAndUseMockNodeManager(mockCtl *gomock.Controller) *MockNodeManager {
	mockNodeManager := NewMockNodeManager(mockCtl)
	registerNodeManager("mockNode", mockNodeManager)
	SetNodeManager("mockNode")
	return mockNodeManager
}
