package cluster

import (
	"sync"

	"github.com/rootyyang/yangkv/pkg/domain/node"
	"github.com/rootyyang/yangkv/pkg/proto"
)

type MasterFailHandler interface {
	MasterFailHandle(pMasterNode node.NodeInterface)
}

type MasterNodeDetector interface {
	//默认MasterNode已经Start()了
	Start(localNodeMeta *proto.NodeMeta, pMasterNode node.NodeInterface) error
	SetNewMaster(pMasterNode node.NodeInterface)
	Stop()
	RegisterDetectMasterFailHandle(MasterFailHandler)
	UnRegisterDetactMasterFailHandle(MasterFailHandler)
}

var gMasterNodeDetector = &defaultMasterNodeDetector{}

func GetMasterNodeDetector() MasterNodeDetector {
	return gMasterNodeDetector
}

type defaultMasterNodeDetector struct {
	masterNode              node.NodeWithHeartBeat
	masterFailHandlers      map[MasterFailHandler]interface{}
	masterFailHandlersMutex sync.RWMutex
}

func (mnd *defaultMasterNodeDetector) Start(localNodeMeta *proto.NodeMeta, pMasterNode node.NodeInterface) error {
	mnd.masterFailHandlers = make(map[MasterFailHandler]interface{})
	mnd.masterNode = node.GetNodeWithHeartBeatProvider().GetNode()
	mnd.masterNode.Start(localNodeMeta, pMasterNode)
	return nil
}

func (mnd *defaultMasterNodeDetector) SetNewMaster(pMasterNode node.NodeInterface) {

}

func (mnd *defaultMasterNodeDetector) RegisterDetectMasterFailHandle(pFailHandler MasterFailHandler) {
	mnd.masterFailHandlersMutex.Lock()
	defer mnd.masterFailHandlersMutex.Unlock()
	mnd.masterFailHandlers[pFailHandler] = nil

}
func (mnd *defaultMasterNodeDetector) UnRegisterDetactMasterFailHandle(pFailHandler MasterFailHandler) {
	mnd.masterFailHandlersMutex.Lock()
	defer mnd.masterFailHandlersMutex.Unlock()
	delete(mnd.masterFailHandlers, pFailHandler)
}

func (mnd *defaultMasterNodeDetector) Stop() {
	mnd.masterNode.Stop()
}
