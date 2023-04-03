package node

import (
	gomock "github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/proto"
)

type PerceivedNodeHeartBeatFail interface {
	HeartBeatFail(node NodeInterface, failTime int)
}

type NodeWithHeartBeat interface {
	Start(localNodeMeta *proto.NodeMeta, pMasterNode NodeInterface) error
	RegisterHeartBeatFailMoreThanThree(PerceivedNodeHeartBeatFail)
	UnRegisterHeartBeatFailMoreThanThree(PerceivedNodeHeartBeatFail)
	Stop()
}

type NodeWithHeartBeatProvider interface {
	GetNode() NodeWithHeartBeat
}
type name2NodeWithHeartBeatProvider struct {
	usedNodeWithHeartBeatProvider  NodeWithHeartBeatProvider
	name2NodeWithHeartBeatProvider map[string]NodeWithHeartBeatProvider
}

var gUseNodeWithHeartBeatProvider NodeWithHeartBeatProvider = new(defaultNodeHeartBeatProvider)

func GetNodeWithHeartBeatProvider() NodeWithHeartBeatProvider {
	return gUseNodeWithHeartBeatProvider
}
func RegisterAndUseMockNodeWithHeartBeat(mockCtl *gomock.Controller) *MockNodeWithHeartBeatProvider {
	mockNodeProvider := NewMockNodeWithHeartBeatProvider(mockCtl)
	gUseNodeWithHeartBeatProvider = mockNodeProvider
	return mockNodeProvider
}
