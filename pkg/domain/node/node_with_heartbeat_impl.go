package node

import (
	"context"
	"sync"
	"time"

	"github.com/rootyyang/yangkv/pkg/infrastructure/system"
	"github.com/rootyyang/yangkv/pkg/proto"
)

type defaultNodeHeartBeatProvider struct {
}

func (*defaultNodeHeartBeatProvider) GetNode() NodeWithHeartBeat {
	return &NodeWithHeartBeatImp{}
}

type NodeWithHeartBeatImp struct {
	localNodeMeta                      *proto.NodeMeta
	remoteNode                         NodeInterface
	cancelContext                      context.Context
	cancelFunc                         context.CancelFunc
	ticker                             *time.Ticker
	heartbeatFailuresNum               int
	perceivedNodeHeartBeatFailMap      map[PerceivedNodeHeartBeatFail]interface{}
	perceivedNodeHeartBeatFailMapMutex sync.RWMutex
}

func (nwhb *NodeWithHeartBeatImp) Start(localNodeMeta *proto.NodeMeta, pMasterNode NodeInterface) error {
	nwhb.localNodeMeta = localNodeMeta
	nwhb.remoteNode = pMasterNode
	nwhb.cancelContext, nwhb.cancelFunc = context.WithCancel(context.Background())
	nwhb.ticker = system.GetClock().Ticker(3 * time.Second)
	nwhb.perceivedNodeHeartBeatFailMap = make(map[PerceivedNodeHeartBeatFail]interface{})
	go nwhb.deamon()
	return nil
}

func (nwhb *NodeWithHeartBeatImp) RegisterHeartBeatFailMoreThanThree(pPerceivedFail PerceivedNodeHeartBeatFail) {
	nwhb.perceivedNodeHeartBeatFailMapMutex.Lock()
	defer nwhb.perceivedNodeHeartBeatFailMapMutex.Unlock()
	nwhb.perceivedNodeHeartBeatFailMap[pPerceivedFail] = nil
}
func (nwhb *NodeWithHeartBeatImp) UnRegisterHeartBeatFailMoreThanThree(pPerceivedFail PerceivedNodeHeartBeatFail) {
	nwhb.perceivedNodeHeartBeatFailMapMutex.Lock()
	defer nwhb.perceivedNodeHeartBeatFailMapMutex.Unlock()
	delete(nwhb.perceivedNodeHeartBeatFailMap, pPerceivedFail)
}

func (nwhb *NodeWithHeartBeatImp) deamon() {
	heartbeatReq := proto.HeartBeatReq{NodeMeta: nwhb.localNodeMeta}
	for {
		select {
		case <-nwhb.cancelContext.Done():
			return
		case <-nwhb.ticker.C:
			ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
			heartbeatRsp, err := nwhb.remoteNode.HeartBeat(ctx, &heartbeatReq)
			if err != nil || heartbeatRsp.RetCode != 0 {
				//TODO log
				nwhb.heartbeatFailuresNum++
				if nwhb.heartbeatFailuresNum > 2 {
					nwhb.perceivedNodeHeartBeatFailMapMutex.RLock()
					for key, _ := range nwhb.perceivedNodeHeartBeatFailMap {
						//TODO 考虑用携协程
						key.HeartBeatFail(nwhb.remoteNode, nwhb.heartbeatFailuresNum)
					}
					nwhb.perceivedNodeHeartBeatFailMapMutex.RUnlock()
				}
			} else {
				nwhb.heartbeatFailuresNum = 0
			}
		}
	}
}

func (nwhb *NodeWithHeartBeatImp) Stop() {
	nwhb.cancelFunc()
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	nwhb.remoteNode.LeaveCluster(ctx, &proto.LeaveClusterReq{NodeMeta: nwhb.localNodeMeta})
	nwhb.remoteNode.Stop()
}
