package consensus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rootyyang/yangkv/pkg/domain/node"
	"github.com/rootyyang/yangkv/pkg/infrastructure/log"
	"github.com/rootyyang/yangkv/pkg/infrastructure/rpc"
	"github.com/rootyyang/yangkv/pkg/infrastructure/system"
)

type raftState int32

const (
	follower = iota
	candidate
	leader
)

type raftLogs struct {
	buffers            []oneRaftLog
	beginPosition      int
	endPosition        int //最后一个日志的后一位
	commitPosition     int
	beginPositionIndex uint64
	nextPositionIndex  uint64
	//lastApplied        uint64 先跟commitindex合并为一，如果是异步apply，则需要lastApplied
	commitIndex  uint64
	stateMachine StateMachine
	logMutex     sync.RWMutex
}

type oneRaftLog struct {
	term          uint64
	command       Command
	commandSerial string
}

func (lc *raftLogs) init(pBufferLen uint32) {
	lc.buffers = make([]oneRaftLog, pBufferLen)
	lc.beginPositionIndex = 0
	lc.nextPositionIndex = 0
}

func (lc *raftLogs) append(pCommand Command, pCurrentTerm uint64) (rMsgIndex uint64, rPrevIndex uint64, rPrevTerm uint64, rErr error) {
	lc.logMutex.Lock()
	defer lc.logMutex.Unlock()
	if (lc.endPosition+1)%len(lc.buffers) == lc.beginPosition {
		rErr = fmt.Errorf("buffers is full")
		return
	}

	var backPostion int
	if lc.endPosition == 0 {
		backPostion = len(lc.buffers) - 1
	} else {
		backPostion = lc.endPosition - 1
	}
	rPrevIndex = lc.nextPositionIndex - 1
	rPrevTerm = lc.buffers[backPostion].term
	rMsgIndex = lc.nextPositionIndex
	lc.buffers[lc.endPosition] = oneRaftLog{pCurrentTerm, pCommand, pCommand.ToString()}
	lc.endPosition = (lc.endPosition + 1) % len(lc.buffers)
	lc.nextPositionIndex++
	return
}

func (lc *raftLogs) upToDateLast(pLogTerm uint64, pLogIndex uint64) bool {
	lastLogIndex, lastLogTerm, isNull := lc.back()
	if isNull {
		return true
	}
	if pLogTerm > lastLogTerm {
		return true
	} else if pLogTerm == lastLogTerm {
		if pLogIndex >= lastLogIndex {
			return true
		}
		return false
	}
	return false
}

func (lc *raftLogs) back() (rLastLogIndex uint64, rLastLogTerm uint64, rIsNull bool) {
	lc.logMutex.RLock()
	defer lc.logMutex.RUnlock()
	if lc.endPosition == lc.beginPosition {
		rIsNull = true
		return
	}
	rLastLogIndex = lc.nextPositionIndex
	rLastLogTerm = lc.buffers[lc.endPosition-1].term
	return
}

func (lc *raftLogs) get(pBeginIndex uint64, pNums int) ([]string, error) {
	lc.logMutex.RLock()
	defer lc.logMutex.RUnlock()
	if pBeginIndex < lc.beginPositionIndex {
		return nil, fmt.Errorf("param smaller than min log index")
	}
	if pBeginIndex >= lc.nextPositionIndex {
		return nil, fmt.Errorf("param bigger than max log index")
	}
	iterBegin := (int(pBeginIndex-lc.beginPositionIndex) + lc.beginPosition) % len(lc.buffers)
	numsInBuffers := lc.nextPositionIndex - pBeginIndex
	if numsInBuffers < uint64(pNums) {
		pNums = int(numsInBuffers)
	}
	rSlice := make([]string, pNums)
	for i := 0; iterBegin != lc.endPosition && i < pNums; i++ {
		rSlice[i] = lc.buffers[iterBegin].commandSerial
	}
	return rSlice, nil
}

//TODO 这里归属于raft对象，是否可以不做判断？
func (lc *raftLogs) commit(pCommitIndex uint64) error {
	lc.logMutex.Lock()
	defer lc.logMutex.Unlock()
	if pCommitIndex <= lc.commitIndex {
		return fmt.Errorf("param smaller than min log index")
	}
	if pCommitIndex >= lc.nextPositionIndex {
		return fmt.Errorf("param bigger than max log index")
	}
	for lc.commitIndex <= pCommitIndex {
		lc.stateMachine.ApplyLog(lc.buffers[lc.commitPosition].command)
		lc.commitPosition++
	}
	lc.commitIndex = pCommitIndex
	return nil
}
func (lc *raftLogs) truncate() error {
	lc.logMutex.Lock()
	defer lc.logMutex.Unlock()
	lc.beginPosition = (lc.commitPosition + 1) % len(lc.buffers)
	lc.beginPositionIndex = lc.commitIndex + 1
	return nil
}

type raft struct {
	currentTerm uint64
	voteFor     string
	state       raftState

	stateMutex sync.RWMutex

	nextIndex              []uint64
	matchIndex             []uint64
	matchAndNextIndexMutex sync.RWMutex
	logs                   *raftLogs

	clusterMajority        int32
	otherClusterNodes      []node.Node
	otherClusterNodesMutex sync.RWMutex

	localNodeMeta node.NodeMeta

	tryRequestVoteTimeMutex sync.RWMutex
	tryRequestVoteTime      time.Time

	masterNodeID string

	/*perceivedMasterMap      map[PerceivedMasterChange]interface{}
	perceivedMasterMapMutex sync.RWMutex*/

	cancelFunc    context.CancelFunc
	cancelContext context.Context

	rpcServer rpc.RPCServer
}

func (rn *raft) Start(pLocalNodeMeta node.NodeMeta, pAllClusterNodes []node.NodeMeta) error {
	rn.otherClusterNodes = make([]node.Node, len(pAllClusterNodes)-1)
	for key, value := range pAllClusterNodes {
		if value.NodeId == pLocalNodeMeta.NodeId {
			continue
		}
		oneNode := node.GetNodeProvider().GetNode(value)
		err := oneNode.Start()
		if err != nil {
			log.GetLogInstance().Errorf("Start RPC Node Error[%v]", err)
		}
		rn.otherClusterNodes[key] = oneNode
	}
	rn.nextIndex = make([]uint64, len(pAllClusterNodes)-1)
	rn.matchIndex = make([]uint64, len(pAllClusterNodes)-1)
	rn.clusterMajority = int32(len(pAllClusterNodes)/2 + 1)
	rn.cancelContext, rn.cancelFunc = context.WithCancel(context.Background())
	rn.state = follower
	rn.localNodeMeta = pLocalNodeMeta
	rpcProvider, err := rpc.GetRPCServerProvider()
	if err != nil || rpcProvider == nil {
		log.GetLogInstance().Errorf("rpc.GetRPCServerProvider()=nil,[%v], want not nil, nil", err)
		return fmt.Errorf("rpcProvider is nil ")
	}
	rn.rpcServer = rpcProvider.GetRPCServer(rn.localNodeMeta.Port)
	if rn.rpcServer == nil {
		log.GetLogInstance().Errorf("rpcProvider.GetRPCServer(%v)=nil, want not nil", pLocalNodeMeta.Port)
		return fmt.Errorf("rpcProvider is nil ")
	}
	err = rn.rpcServer.RegisterRaftHandle(rn)
	if err != nil {
		log.GetLogInstance().Errorf("rpcServer.RegisterRaftHandle=[%v], want nil", err)
		return err
	}
	err = rn.rpcServer.Start()
	if err != nil {
		log.GetLogInstance().Errorf("rpcServer.Start()=[%v], want nil", err)
		return err
	}
	rn.logs = new(raftLogs)
	rn.logs.init(1000)
	go rn.tickHeartBeat()
	go rn.tickTryRequestVote()
	log.GetLogInstance().Debugf("node[%v] raft start", rn.localNodeMeta)
	return nil
}

func (rn *raft) Stop() {
	rn.rpcServer.Stop()
	rn.cancelFunc()
	rn.stateMutex.Lock()
	rn.state = follower
	rn.stateMutex.Unlock()
	for _, value := range rn.otherClusterNodes {
		value.Stop()
	}
}

/*
服务启动以后，先探测nextIndex避免
超时和失败重试，应该在状态机中进行

这里数据同步的几种方式，主要注意乱序和丢失的问题，并且在节点返回失败之后，要移动nextIndex
解决方案(当前选择方案三):
一.每个节点由单个协程同步数据，为了解决QPS问题，一次同步，应该是包含多条日志
问题1: 延时，用户的延时可能是两轮同步的网络延时，即一个新的请求到了，需要等待上一轮的同步结束
优点1: 实现简单
二.每个请求，由一个协程同步一次数据，并且负责之后的重试，(该方案不靠谱)
问题1: 遇到返回false应该如何处理？
问题2: 如果单个节点失败，会导致协程数不断累计
三.每个请求，由一个协程同步一次数据，失败后，由单个协程负责重试(该方案，正常情况下延时最低)
问题1: 协程应该如何负责重试？ 收到一次失败通知，批量重试一次
问题2: 重试策略？
*/
func (rn *raft) ApplyMsg(ctx context.Context, pCommand Command) error {
	rn.stateMutex.RLock()
	tmpCurrentTerm := rn.currentTerm
	tmpCurrentStat := rn.state
	rn.stateMutex.RUnlock()
	if tmpCurrentStat != leader {
		return fmt.Errorf("Node Not Leader")
	}
	msgIndex, prevIndex, prevTerm, err := rn.logs.append(pCommand, rn.currentTerm)
	if err != nil {
		return err
	}
	//等待ctx超时
	syncResultChannel := make(chan bool)
	//同步数据
	rn.trySyncMsgAndHandleResult(tmpCurrentTerm, msgIndex, []string{pCommand.ToString()}, prevIndex, prevTerm)
	select {
	case <-ctx.Done():
		return fmt.Errorf("timeout Error")
	case success := <-syncResultChannel:
		if success {
			return nil
		} else {
			return fmt.Errorf("sync follower error")
		}
	}
}

//同步有两种方式
//定义一些全局变量，每个协程都尝试抢锁，然后做状态更新
//通过channel，交给一个统一的协程进行处理
func (rn *raft) trySyncMsgAndHandleResult(pSuccessResult chan<- bool, pCurrentTerm uint64, pCurrentIndex uint64, pSerialCommand []string, pPrevIndex uint64, pPrevTerm uint64) error {
	type appendEntriesRespWrapper struct {
		key  int
		resp *rpc.AppendEntiresResp
	}
	AppendRespChannel := make(chan appendEntriesRespWrapper)
	for key, value := range rn.otherClusterNodes {
		tmpKey := key
		tmpValue := value
		ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
		resp, err := tmpValue.GetRPC().AppendEntires(ctx, &rpc.AppendEntiresReq{LeaderTerm: pCurrentTerm, LeaderId: rn.localNodeMeta.NodeId, PrevLogIndex: pPrevIndex, PrevLogTerm: pPrevTerm})
		//每个都要做的，在自己线程中做
		if err != nil {
			AppendRespChannel <- appendEntriesRespWrapper{tmpKey, nil}

		} else {
			AppendRespChannel <- appendEntriesRespWrapper{tmpKey, resp}
		}
	}
	otherNodeNumber := len(rn.otherClusterNodes)
	successKeys := make([]int, 0, otherNodeNumber)
	var maxRespTerm uint64
	for i := 0; i < otherNodeNumber; i++ {
		oneResp := <-AppendRespChannel
		if oneResp.resp == nil {
			continue
		}
		if oneResp.resp.Term > maxRespTerm {
			maxRespTerm = oneResp.resp.Term
		}
		if oneResp.resp.Success {
			successKeys = append(successKeys, oneResp.key)
		}
	}
	if maxRespTerm > pCurrentTerm {
		//更新节点为follower
		rn.stateMutex.Lock()
		if maxRespTerm > rn.currentTerm {
			rn.currentTerm = maxRespTerm
			rn.state = follower
			rn.voteFor = ""
			rn.stateMutex.Unlock()
			rn.setNextRetryVote()
			return fmt.Errorf("Node Not Leader")
		}
		rn.stateMutex.Unlock()
		return fmt.Errorf("Please retry again")
	}
	//尝试更新next Index
	rn.matchAndNextIndexMutex.Lock()
	for key := range successKeys {
		if rn.nextIndex[key] <= pCurrentIndex {
			rn.nextIndex[key] = pCurrentIndex
		}
	}
	rn.matchAndNextIndexMutex.Unlock()
	//尝试更新
	if len(successKeys)+1 >= int(rn.clusterMajority) {
		return rn.logs.commit(pCurrentIndex)
	}
	//触发重试
	return fmt.Errorf("Not Sync Majority")
}

/*
func (rn *raft) RegisterPerceivedMasterChange(pPMC PerceivedMasterChange) {
	rn.perceivedMasterMapMutex.Lock()
	defer rn.perceivedMasterMapMutex.Unlock()
	rn.perceivedMasterMap[pPMC] = nil
}

func (rn *raft) UnRegisterPerceivedMasterChange(pPMC PerceivedMasterChange) {
	rn.perceivedMasterMapMutex.Lock()
	defer rn.perceivedMasterMapMutex.Unlock()
	delete(rn.perceivedMasterMap, pPMC)
}*/

func (rn *raft) persistent() error {
	return nil
}

func (rn *raft) AppendEntires(ctx context.Context, req *rpc.AppendEntiresReq) (*rpc.AppendEntiresResp, error) {
	/*rn.stateMutex.RLock()
	tmpCurrentTerm := rn.currentTerm
	tmpCurrentState := rn.state
	rn.stateMutex.RUnlock()

	//最常见的情况，用读锁判断
	if req.LeaderTerm == tmpCurrentTerm && tmpCurrentState == follower {
		rn.matchAndNextIndexMutex.Lock()
		defer rn.matchAndNextIndexMutex.Unlock()
		if len(rn.log)-1 < req.PrevLogIndex || rn.log[req.PrevLogIndex].term != req.PrevLogTerm {
			return &rpc.AppendEntiresResp{Term: tmpCurrentTerm, Success: false}, nil
		}
		for key, value := range req.Entries {

		}

		//心跳，直接返回false
		if len(req.Entries) == 0 {
			rn.setNextRetryVote()
			return &rpc.AppendEntiresResp{Term: tmpCurrentTerm, Success: true}, nil
		}
		//非心跳，

	}
	//这种情况直接返回false
	if req.LeaderTerm < tmpCurrentTerm {
		return &rpc.AppendEntiresResp{Term: rn.currentTerm, Success: false}, nil
	}
	//低概率情况，直接加写锁，二次检查各种状态
	rn.stateMutex.Lock()
	defer rn.stateMutex.Unlock()
	if req.LeaderTerm > rn.currentTerm {
		log.GetLogInstance().Debugf("node[%v] update to term[%v] in AppendEntires()", rn.localNodeMeta.NodeId, req.LeaderTerm)
		rn.currentTerm = req.LeaderTerm
		rn.state = follower
		rn.voteFor = ""
		//刷新时钟
		rn.setNextRetryVote()
		return &rpc.AppendEntiresResp{Term: rn.currentTerm, Success: true}, nil
	} else if req.LeaderTerm == tmpCurrentTerm {
		rn.state = follower
		//刷新时钟
		rn.setNextRetryVote()
		return &rpc.AppendEntiresResp{Term: rn.currentTerm, Success: true}, nil
	}
	//不可能走到这里，rn.currentTerm不可能降低
	log.GetLogInstance().Errorf("tmpCurrentTerm[%v] > rn.curentTerm[%v]", tmpCurrentTerm, rn.currentTerm)
	return &rpc.AppendEntiresResp{Term: rn.currentTerm, Success: false}, nil*/
	return nil, nil
}
func (rn *raft) RequestVote(ctx context.Context, req *rpc.RequestVoteReq) (*rpc.RequestVoteResp, error) {
	//RequestVote理论上应该是小概率事件，所以直接全局加锁
	rn.stateMutex.Lock()
	defer rn.stateMutex.Unlock()
	log.GetLogInstance().Debugf("node{id:%v, term:%v} get request vote from node{id:%v, term:%v }", rn.localNodeMeta.NodeId, rn.currentTerm, req.CandidateId, req.Term)
	if req.Term < rn.currentTerm {
		return &rpc.RequestVoteResp{Term: rn.currentTerm, VoteGranted: false}, nil
	} else if req.Term == rn.currentTerm {
		if rn.state == follower && (rn.voteFor == "" || rn.voteFor == req.CandidateId) && rn.logs.upToDateLast(req.LastLogTerm, req.LastLogIndex) {
			rn.voteFor = req.CandidateId
			//刷新时钟
			rn.setNextRetryVote()
			return &rpc.RequestVoteResp{Term: rn.currentTerm, VoteGranted: true}, nil
		}
		return &rpc.RequestVoteResp{Term: rn.currentTerm, VoteGranted: false}, nil
	}
	log.GetLogInstance().Debugf("node[%v] update to term [%v] in RequestVote()", rn.localNodeMeta.NodeId, req.Term)
	rn.currentTerm = req.Term
	rn.state = follower
	if rn.logs.upToDateLast(req.LastLogTerm, req.LastLogIndex) {
		rn.voteFor = req.CandidateId
		rn.setNextRetryVote()
		return &rpc.RequestVoteResp{Term: rn.currentTerm, VoteGranted: true}, nil
	}
	return &rpc.RequestVoteResp{Term: rn.currentTerm, VoteGranted: false}, nil
}

func (rn *raft) tickHeartBeat() {
	heartBeatTicker := system.GetClock().Ticker(20 * time.Millisecond)
	for {
		select {
		case <-rn.cancelContext.Done():
			heartBeatTicker.Stop()
			return
		case <-heartBeatTicker.GetChannel():
			rn.stateMutex.RLock()
			tmpCurrentStat := rn.state
			tmpCurrentTerm := rn.currentTerm
			rn.stateMutex.RUnlock()
			if tmpCurrentStat != leader {
				break
			}
			maxRespTerm := rn.sendHeartBeat(tmpCurrentTerm)
			//最常见的情况
			if maxRespTerm <= tmpCurrentTerm {
				//更新时钟
				break
			}
			rn.stateMutex.Lock()
			if maxRespTerm > rn.currentTerm {
				log.GetLogInstance().Debugf("node[%v] update to term [%v] in tickHeartBeat()", rn.localNodeMeta.NodeId, maxRespTerm)
				rn.currentTerm = maxRespTerm
				rn.state = follower
				rn.voteFor = ""
				//更新时钟为随机值
				rn.stateMutex.Unlock()
				rn.setNextRetryVote()
				break
			}
			//else不做更新，证明已经有其他的心跳，或者操作，更新过时钟
			rn.stateMutex.Unlock()

		}
	}

}

func (rn *raft) tickTryRequestVote() {
	firstRetryTime := rn.setNextRetryVote()
	sleepTimer := system.GetClock().Timer(time.Duration(firstRetryTime) * time.Millisecond)
	for {
		select {
		case <-rn.cancelContext.Done():
			sleepTimer.Stop()
			return
		case now := <-sleepTimer.GetChannel():
			rn.stateMutex.RLock()
			tmpCurrentStat := rn.state
			rn.stateMutex.RUnlock()
			if tmpCurrentStat == leader {
				break
			}
			rn.tryRequestVoteTimeMutex.RLock()
			nexTryTime := rn.tryRequestVoteTime
			rn.tryRequestVoteTimeMutex.RUnlock()
			if nexTryTime.After(now) {
				sleepTimer.Reset(nexTryTime.Sub(now))
				continue
			}

			rn.stateMutex.Lock()
			rn.state = candidate
			rn.currentTerm++
			rn.voteFor = rn.localNodeMeta.NodeId
			tmpCurrentTerm := rn.currentTerm
			rn.stateMutex.Unlock()
			log.GetLogInstance().Debugf("node[%v] begin request vote term[%v] time[%v]", rn.localNodeMeta.NodeId, tmpCurrentTerm, now.UnixMilli())
			maxRespTerm, voteNum := rn.sendRequstVote(tmpCurrentTerm)
			//如果收到一个
			if maxRespTerm > tmpCurrentTerm {
				rn.stateMutex.Lock()
				if maxRespTerm > rn.currentTerm {
					rn.currentTerm = maxRespTerm
					rn.state = follower
					rn.voteFor = ""
				}
				rn.stateMutex.Unlock()
				//这里证明在心跳的时候，或者投票的时候，已经更新了rn.currentTerm和rn.stat
				break
			}
			log.GetLogInstance().Debugf("node[%v] vote num[%v]", rn.localNodeMeta.NodeId, voteNum)
			if voteNum >= rn.clusterMajority {
				rn.stateMutex.Lock()
				if rn.currentTerm == tmpCurrentTerm {
					//优化启停
					rn.state = leader
					log.GetLogInstance().Debugf("node[%v] become leader", rn.localNodeMeta.NodeId)
				}
				rn.stateMutex.Unlock()
				//这里证明在心跳的时候，或者投票的时候，已经更新了rn.currentTerm和rn.stat
			}
		}
		nextSleep := rn.setNextRetryVote()
		sleepTimer.Reset(time.Duration(nextSleep) * time.Millisecond)
	}

}

func (rn *raft) sendHeartBeat(pCurrentTerm uint64) uint64 {
	heartBeatRespChannel := make(chan *rpc.AppendEntiresResp)
	rn.otherClusterNodesMutex.RLock()
	otherNodeNumber := len(rn.otherClusterNodes)
	for _, value := range rn.otherClusterNodes {
		tmpValue := value
		go func() {
			ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
			resp, error := tmpValue.GetRPC().AppendEntires(ctx, &rpc.AppendEntiresReq{LeaderTerm: pCurrentTerm, LeaderId: rn.localNodeMeta.NodeId})
			if error != nil {
				heartBeatRespChannel <- nil
			} else {
				heartBeatRespChannel <- resp
			}
		}()
	}
	rn.otherClusterNodesMutex.RUnlock()
	maxRespTerm := pCurrentTerm
	for i := 0; i < otherNodeNumber; i++ {
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
func (rn *raft) sendRequstVote(pCurrentTerm uint64) (uint64, int32) {
	voteChannel := make(chan *rpc.RequestVoteResp)
	for _, value := range rn.otherClusterNodes {
		//如果没有tmpNode，会导致多个
		tmpNode := value
		go func() {
			ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
			resp, err := tmpNode.GetRPC().RequestVote(ctx, &rpc.RequestVoteReq{Term: pCurrentTerm, CandidateId: rn.localNodeMeta.NodeId})
			if err != nil {
				log.GetLogInstance().Debugf("node[%v] RequestVote(%v) error[%v]", rn.localNodeMeta.NodeId, tmpNode.GetNodeMeta(), err)
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
		log.GetLogInstance().Debugf("node[%v] get request vote resp[%v]", rn.localNodeMeta.NodeId, *oneResp)
		if oneResp.VoteGranted == true {
			voteNum++
		}
		if oneResp.Term > maxRespTerm {
			maxRespTerm = oneResp.Term
		}
	}
	return maxRespTerm, voteNum
}

/*
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
}*/

func (rn *raft) GetRaftState() (raftState, uint64) {
	rn.stateMutex.RLock()
	defer rn.stateMutex.RUnlock()
	return rn.state, rn.currentTerm
}

func (rn *raft) setNextRetryVote() int16 {
	ms := int16(150 + (system.GetRandom().Int() % 151))
	now := system.GetClock().Now()
	tryRequestVoteTime := now.Add(time.Duration(ms) * time.Millisecond)
	rn.tryRequestVoteTimeMutex.Lock()
	defer rn.tryRequestVoteTimeMutex.Unlock()
	if rn.tryRequestVoteTime.Before(tryRequestVoteTime) {
		rn.tryRequestVoteTime = tryRequestVoteTime
		return ms
	}
	return int16(rn.tryRequestVoteTime.Sub(tryRequestVoteTime))

}
