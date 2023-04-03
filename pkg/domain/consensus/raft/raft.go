package raft

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/rootyyang/yangkv/pkg/domain/consensus"
	"github.com/rootyyang/yangkv/pkg/domain/node"
	"github.com/rootyyang/yangkv/pkg/infrastructure/log"
	"github.com/rootyyang/yangkv/pkg/infrastructure/rpc"
)

type raftState int32

const (
	follower = iota
	candidate
	leader
)

type raft struct {
	currentTerm int32
	voteFor     string
	state       raftState

	stateMutex sync.RWMutex

	clusterMajority int32

	otherClusterNodes map[string]node.NodeInterface

	localNodeMeta *node.NodeMeta
	masterNodeID  string

	perceivedMasterMap      map[consensus.PerceivedMasterChange]interface{}
	perceivedMasterMapMutex sync.RWMutex

	cancelFunc        context.CancelFunc
	cancelContext     context.Context
	timer             *time.Timer
	resetTimerChannal chan int16
}

func (rn *raft) Start(pLocalNodeMeta *node.NodeMeta, pOtherClusterNodes []node.NodeInterface) error {
	for _, value := range pOtherClusterNodes {
		rn.otherClusterNodes[value.GetNodeMeta().NodeId] = value
	}
	rn.clusterMajority = int32(len(pOtherClusterNodes)/2 + 1)
	rn.cancelContext, rn.cancelFunc = context.WithCancel(context.Background())
	rn.state = follower

	ms := 150 + (rand.Int() % 150)
	rn.timer = time.NewTimer(time.Duration(ms) * time.Microsecond)
	rn.resetTimerChannal = make(chan int16)
	go rn.daemon()
	return nil
}

func (rn *raft) Stop() {

}

func (rn *raft) RegisterPerceivedMasterChange(pPMC consensus.PerceivedMasterChange) {
	rn.perceivedMasterMapMutex.Lock()
	defer rn.perceivedMasterMapMutex.Unlock()
	rn.perceivedMasterMap[pPMC] = nil
}

func (rn *raft) UnRegisterPerceivedMasterChange(pPMC consensus.PerceivedMasterChange) {
	rn.perceivedMasterMapMutex.Lock()
	defer rn.perceivedMasterMapMutex.Unlock()
	delete(rn.perceivedMasterMap, pPMC)
}

func (rn *raft) persistent() error {
	return nil
}

func (rn *raft) AppendEntires(ctx context.Context, req *rpc.AppendEntiresReq) (*rpc.AppendEtriesResp, error) {
	rn.stateMutex.RLock()
	tmpCurrentTerm := rn.currentTerm
	tmpCurrentState := rn.state
	rn.stateMutex.RUnlock()
	//最常见的情况，用读锁判断
	if tmpCurrentTerm == tmpCurrentTerm && tmpCurrentState == follower {
		//更新时钟
		rn.resetTimer(true)
		return &rpc.AppendEtriesResp{Term: rn.currentTerm, Success: true}, nil
	}
	//这种情况直接返回false
	if req.LeaderTerm < tmpCurrentTerm {
		return &rpc.AppendEtriesResp{Term: rn.currentTerm, Success: false}, nil
	}
	//其他情况都需要更新状态
	rn.stateMutex.Lock()
	defer rn.stateMutex.Unlock()
	if req.LeaderTerm > rn.currentTerm {
		rn.currentTerm = req.LeaderTerm
		rn.state = follower
		rn.voteFor = ""
		//刷新时钟
		rn.resetTimer(true)
		return &rpc.AppendEtriesResp{Term: rn.currentTerm, Success: true}, nil
	} else if req.LeaderTerm == tmpCurrentTerm {
		rn.state = follower
		//刷新时钟
		rn.resetTimer(true)
		return &rpc.AppendEtriesResp{Term: rn.currentTerm, Success: true}, nil
	}
	//不可能走到这里，rn.currentTerm不可能降低
	log.GetLogInstance().Errorf("tmpCurrentTerm[%v] > rn.cuurentTerm[%v]", tmpCurrentTerm, rn.currentTerm)
	return &rpc.AppendEtriesResp{Term: rn.currentTerm, Success: false}, nil

}
func (rn *raft) RequestVote(ctx context.Context, req *rpc.RequestVoteReq) (*rpc.RequestVoteResp, error) {
	//RequestVote理论上应该是小概率事件，所以直接全局加锁
	rn.stateMutex.Lock()
	defer rn.stateMutex.Unlock()
	if req.Term < rn.currentTerm {
		return &rpc.RequestVoteResp{Term: rn.currentTerm, VoteGranted: false}, nil
	} else if req.Term == rn.currentTerm {
		if rn.state == follower && rn.voteFor == "" {
			rn.voteFor = req.CandidateId
			//刷新时钟
			rn.resetTimer(true)
			return &rpc.RequestVoteResp{Term: rn.currentTerm, VoteGranted: true}, nil
		}
		return &rpc.RequestVoteResp{Term: rn.currentTerm, VoteGranted: false}, nil
	}
	rn.currentTerm = req.Term
	rn.state = follower
	rn.voteFor = req.CandidateId
	//刷新时钟
	rn.resetTimer(true)
	return &rpc.RequestVoteResp{Term: rn.currentTerm, VoteGranted: true}, nil
}

func (rn *raft) daemon() {
	for {
		select {
		case <-rn.cancelContext.Done():
			return
		case <-rn.timer.C:
			rn.stateMutex.RLock()
			tmpCurrentStat := rn.state
			tmpCurrentTerm := rn.currentTerm
			rn.stateMutex.RUnlock()
			if tmpCurrentStat == leader {
				//sendHeartBeat
				maxRespTerm := rn.sendHeartBeat(tmpCurrentTerm)
				//最常见的情况
				if maxRespTerm <= tmpCurrentTerm {
					//更新时钟
					rn.resetTimer(false)
					break
				}
				rn.stateMutex.Lock()
				if maxRespTerm > rn.currentTerm {
					rn.currentTerm = maxRespTerm
					rn.state = follower
					//更新时钟为随机值
					rn.resetTimer(true)
				}
				rn.stateMutex.Unlock()
				//如果currentTerm增加了，只可能变成follower什么都不做
				break

			}
			rn.stateMutex.Lock()
			rn.state = candidate
			rn.currentTerm++
			rn.voteFor = rn.localNodeMeta.NodeId
			tmpCurrentTerm = rn.currentTerm
			rn.stateMutex.Unlock()
			maxRespTerm, voteNum := rn.sendRequstVote(tmpCurrentTerm)
			//如果收到一个
			if maxRespTerm > tmpCurrentTerm {
				rn.stateMutex.Lock()
				if maxRespTerm > rn.currentTerm {
					rn.currentTerm = maxRespTerm
					rn.state = follower
					rn.voteFor = ""
					//刷新时钟
					rn.resetTimer(true)
				}
				rn.stateMutex.Unlock()
				//这里证明在心跳的时候，或者投票的时候，已经更新了rn.currentTerm和rn.stat
				break
			}
			if voteNum > rn.clusterMajority {
				rn.stateMutex.Lock()
				if rn.currentTerm == tmpCurrentTerm {
					rn.state = leader
					//刷新时钟为固定值
					rn.resetTimer(false)
				}
				rn.stateMutex.Unlock()
				//这里证明在心跳的时候，或者投票的时候，已经更新了rn.currentTerm和rn.stat
				break
			}
			//没有获得足够的投票
			//刷新时钟为随机值
			rn.resetTimer(true)
		case newms := <-rn.resetTimerChannal:
			rn.timer.Stop()
			for len(rn.timer.C) > 0 {
				<-rn.timer.C
			}
			rn.timer.Reset(time.Duration(newms) * time.Millisecond)
		}
	}
	return
}
func (rn *raft) resetTimer(random bool) {
	var ms int16
	if random {
		ms = int16(150 + (rand.Int() % 150))
	} else {
		ms = 30
	}
	rn.resetTimerChannal <- int16(ms)
}

func (rn *raft) sendHeartBeat(pCurrentTerm int32) int32 {
	heartBeatRespChannel := make(chan *rpc.AppendEtriesResp)
	for _, value := range rn.otherClusterNodes {
		go func() {
			ctx, _ := context.WithTimeout(context.Background(), 100*time.Microsecond)
			resp, error := value.GetRPC().AppendEntires(ctx, &rpc.AppendEntiresReq{pCurrentTerm, rn.localNodeMeta.NodeId})
			if error != nil {
				heartBeatRespChannel <- nil
			} else {
				heartBeatRespChannel <- resp
			}
		}()
	}
	maxRespTerm := pCurrentTerm
	for i := 0; i < len(rn.otherClusterNodes); i++ {
		oneResp := <-heartBeatRespChannel
		if oneResp == nil {
			continue
		}
		if oneResp.Term > maxRespTerm {
			maxRespTerm = oneResp.Term
		}
	}
	return maxRespTerm
}
func (rn *raft) sendRequstVote(pCurrentTerm int32) (int32, int32) {
	voteChannel := make(chan *rpc.RequestVoteResp)
	for _, value := range rn.otherClusterNodes {
		go func() {
			ctx, _ := context.WithTimeout(context.Background(), 100*time.Microsecond)
			resp, error := value.GetRPC().RequestVote(ctx, &rpc.RequestVoteReq{pCurrentTerm, rn.localNodeMeta.NodeId})
			if error != nil {
				voteChannel <- nil
			} else {
				voteChannel <- resp
			}
		}()
	}
	maxRespTerm := pCurrentTerm
	var voteNum int32 = 1
	for i := 0; i < len(rn.otherClusterNodes); i++ {
		oneResp := <-voteChannel
		if oneResp == nil {
			continue
		}
		if oneResp.VoteGranted == true {
			voteNum++
		}
		if oneResp.Term > maxRespTerm {
			maxRespTerm = oneResp.Term
		}
	}
	return maxRespTerm, voteNum
}

func (rn *raft) notifyBecomeMaster() {
	rn.perceivedMasterMapMutex.RLock()
	defer rn.perceivedMasterMapMutex.RUnlock()
	for key, _ := range rn.perceivedMasterMap {
		key.NoLongerMaster()
	}
}

func (rn *raft) notifyNolongerMaster() {
	rn.perceivedMasterMapMutex.RLock()
	defer rn.perceivedMasterMapMutex.RUnlock()
	for key, _ := range rn.perceivedMasterMap {
		key.BecomeMaster()
	}
}
