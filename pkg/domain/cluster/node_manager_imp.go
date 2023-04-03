package cluster

import (
	"context"
	"sync"
	"time"

	"github.com/rootyyang/yangkv/pkg/domain/node"
	"github.com/rootyyang/yangkv/pkg/infrastructure/log"
	"github.com/rootyyang/yangkv/pkg/infrastructure/system"
	"github.com/rootyyang/yangkv/pkg/proto"
)

func init() {
	registerNodeManager("default", &defaultNodeManager{})
}

type defaultNodeManager struct {
	localNode     *proto.NodeMeta
	nodeMap       map[string]*nodeWithHeartBeat
	mapLock       sync.RWMutex
	cancelFunc    context.CancelFunc
	cancelContext context.Context
	ticker        *time.Ticker
}

type nodeWithHeartBeat struct {
	remoteNode node.NodeInterface
	//该变量主要用于心跳
	localNode            *proto.NodeMeta
	cancelContext        context.Context
	ticker               *time.Ticker
	heartbeatMutex       sync.RWMutex
	heartbeatFailuresNum int32
}

func (nwhb *nodeWithHeartBeat) GetHeartBeatFailuresNum() int32 {
	nwhb.heartbeatMutex.RLock()
	defer nwhb.heartbeatMutex.RUnlock()
	return nwhb.heartbeatFailuresNum
}

func (nwhb *nodeWithHeartBeat) Start(pLocalNode *proto.NodeMeta, pRemoteNode node.NodeInterface, pCancelContext context.Context) error {
	nwhb.localNode = pLocalNode
	nwhb.remoteNode = pRemoteNode
	nwhb.cancelContext = pCancelContext
	if err := nwhb.remoteNode.Start(); err != nil {
		return err
	}
	nwhb.ticker = system.GetClock().Ticker(3 * time.Second)
	go nwhb.daemon()
	return nil
}

func (nwhb *nodeWithHeartBeat) Stop() error {
	nwhb.ticker.Stop()
	nwhb.remoteNode.Stop()
	return nil
}

func (nwhb *nodeWithHeartBeat) daemon() {
	//proto.HeartBeatReq
	nodeMeta := nwhb.localNode
	heartbeatReq := proto.HeartBeatReq{NodeMeta: nodeMeta}
	for {
		select {
		case <-nwhb.cancelContext.Done():
			return
		case <-nwhb.ticker.C:
			ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
			heartbeatRsp, err := nwhb.remoteNode.HeartBeat(ctx, &heartbeatReq)
			nwhb.heartbeatMutex.Lock()
			if err != nil || heartbeatRsp.RetCode != 0 {
				//TODO log
				nwhb.heartbeatFailuresNum++
			} else {
				nwhb.heartbeatFailuresNum = 0
			}
			nwhb.heartbeatMutex.Unlock()
		}
	}
}

func (nm *defaultNodeManager) Start() error {
	nm.ticker = system.GetClock().Ticker(3 * time.Second)
	go nm.heartBeatDaemon()
	return nil
}

func (nm *defaultNodeManager) Stop() error {
	nm.ticker.Stop()
	nm.cancelFunc()
	nm.mapLock.Lock()
	defer nm.mapLock.Unlock()
	for _, value := range nm.nodeMap {
		value.Stop()
	}
	return nil
}

func (nm *defaultNodeManager) RegisterBeforeJoinFunc() error {
	return nil
}

func (nm *defaultNodeManager) UnregisterBeforeJoinFunc() error {
	return nil
}

func (nm *defaultNodeManager) RegisterAfterLeftFunc() error {
	return nil
}

func (nm *defaultNodeManager) UnregisterAfterLeftFunc() error {
	return nil
}

func (nm *defaultNodeManager) JoinCluster(pNodeMeta *proto.NodeMeta) error {
	nm.mapLock.RLock()
	if _, ok := nm.nodeMap[pNodeMeta.NodeId]; ok {
		nm.mapLock.RUnlock()
		return nil
	}
	nm.mapLock.RUnlock()
	nodeProvider, err := node.GetNodeProvider()
	if err != nil {
		return err
	}
	remoteNode := nodeProvider.GetNode(pNodeMeta)
	nodeWithHearbeat := nodeWithHeartBeat{}
	if err := nodeWithHearbeat.Start(nm.localNode, remoteNode, nm.cancelContext); err != nil {
		return err
	}
	nm.mapLock.Lock()
	nm.nodeMap[pNodeMeta.NodeId] = &nodeWithHearbeat
	nm.mapLock.Unlock()
	log.GetLogInstance().Debugf("IP[%v], Port[%v] join", pNodeMeta.Ip, pNodeMeta.Port)
	return nil
}

func (nm *defaultNodeManager) LeftCluster(pNodeMeta *proto.NodeMeta) error {
	nm.mapLock.Lock()
	cn, ok := nm.nodeMap[pNodeMeta.NodeId]
	if !ok {
		nm.mapLock.Unlock()
		return nil
	}
	delete(nm.nodeMap, pNodeMeta.NodeId)
	nm.mapLock.Unlock()
	cn.Stop()
	log.GetLogInstance().Debugf("NodeIP[%v], Port[%v] end left", pNodeMeta.Ip, pNodeMeta.Port)
	return nil
}
func (nm *defaultNodeManager) heartBeatDaemon() {
	for {
		select {
		case <-nm.cancelContext.Done():
			return
		case <-nm.ticker.C:
			wait_delete_keys := []*proto.NodeMeta{}
			nm.mapLock.RLock()
			for _, value := range nm.nodeMap {
				if value.GetHeartBeatFailuresNum() > 2 {
					wait_delete_keys = append(wait_delete_keys, value.remoteNode.GetNodeMeta())
				}
			}
			nm.mapLock.RUnlock()
			for _, value := range wait_delete_keys {
				nm.LeftCluster(value)
			}
		}
	}
	return
}

func (nm *defaultNodeManager) GetAllNode() map[string]node.NodeInterface {
	retMap := make(map[string]node.NodeInterface)
	nm.mapLock.RLock()
	defer nm.mapLock.RUnlock()
	for key, value := range nm.nodeMap {
		retMap[key] = value.remoteNode
	}
	return retMap

}
func (nm *defaultNodeManager) GetNodeByID(pID string) node.NodeInterface {
	return nil
}

func (nm *defaultNodeManager) Init(localNode *proto.NodeMeta) {
	nm.localNode = localNode
	nm.nodeMap = make(map[string]*nodeWithHeartBeat)
	nm.cancelContext, nm.cancelFunc = context.WithCancel(context.Background())
	return
}
