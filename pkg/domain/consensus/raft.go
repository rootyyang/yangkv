package consensus

import (
	"container/heap"
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
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

//TODO 之后换成泛型实现
type appendEntiresReqWithRespChannel struct {
	req  *rpc.AppendEntiresReq
	resp chan error
}

type priorityQueue struct {
	buffers []appendEntiresReqWithRespChannel
}

func (pq *priorityQueue) Len() int {
	return len(pq.buffers)
}
func (pq *priorityQueue) Less(i, j int) bool {
	return pq.buffers[i].req.PrevLogIndex < pq.buffers[j].req.PrevLogIndex
}
func (pq *priorityQueue) Swap(i, j int) {
	pq.buffers[i], pq.buffers[j] = pq.buffers[j], pq.buffers[i]
}
func (pq *priorityQueue) Push(x any) {
	item := x.(appendEntiresReqWithRespChannel)
	pq.buffers = append(pq.buffers, item)
}
func (pq *priorityQueue) Pop() any {
	old := pq.buffers
	n := len(old)
	x := old[n-1]
	pq.buffers = old[0 : n-1]
	return x
}
func (pq *priorityQueue) Top() any {
	return pq.buffers[len(pq.buffers)]
}

type raftConfig struct {
	heartBeatIntervalMs  int
	retryBatchSyncLogNum int
	retryIntervalMs      int

	logBufferLen     int
	rpcTimeoutTimeMs int
}

type retryAsyncLogWrapper struct {
	index        uint64
	termNotMatch bool
	respChannel  chan bool
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
	otherNodeRetryChannel  []chan retryAsyncLogWrapper
	otherClusterNodesMutex sync.RWMutex

	localNodeMeta node.NodeMeta

	waitInsertQueue *priorityQueue

	tryRequestVoteTimeMutex sync.RWMutex
	tryRequestVoteTime      time.Time

	masterNodeID string

	config raftConfig

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
			log.Instance().Errorf("Start RPC Node Error[%v]", err)
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
		log.Instance().Errorf("rpc.GetRPCServerProvider()=nil,[%v], want not nil, nil", err)
		return fmt.Errorf("rpcProvider is nil ")
	}
	rn.rpcServer = rpcProvider.GetRPCServer(rn.localNodeMeta.Port)
	if rn.rpcServer == nil {
		log.Instance().Errorf("rpcProvider.GetRPCServer(%v)=nil, want not nil", pLocalNodeMeta.Port)
		return fmt.Errorf("rpcProvider is nil ")
	}
	err = rn.rpcServer.RegisterRaftHandle(rn)
	if err != nil {
		log.Instance().Errorf("rpcServer.RegisterRaftHandle=[%v], want nil", err)
		return err
	}
	err = rn.rpcServer.Start()
	if err != nil {
		log.Instance().Errorf("rpcServer.Start()=[%v], want nil", err)
		return err
	}
	rn.logs = new(raftLogs)
	rn.logs.init(1000)
	go rn.tickHeartBeat()
	go rn.tickTryRequestVote()
	log.Instance().Debugf("node[%v] raft start", rn.localNodeMeta)
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
实现:  可以将每个index对应channel保存下来，当同步多数派以后，直接通过channel给用户返回成功
二.每个请求，由一个协程同步一次数据，并且负责之后的重试，(该方案不靠谱)
问题1: 遇到返回false应该如何处理？
问题2: 如果单个节点失败，会导致协程数不断累计
三.每个请求，由一个协程同步一次数据，失败后，由单个协程负责重试，(该方案，正常情况下延时最低)
问题1: 协程应该如何负责重试？ 收到一次失败通知，批量重试一次
问题2: 重试策略？
问题3: 如果重试马上成功了，如何给用户返回成功(这里重试的时候？)
问题4: 是每个节点一个充实协程，还是所有节点一个重试协程？
4.1 每个节点一个重试协程？节点失败以后，就触发重试，同步至majority的节点，负责给用户回复成功
给多个节点同步有两种方式
1.定义一些全局变量，每个协程都尝试抢锁，然后做状态更新
2.通过channel，在统一的逻辑进行处理
*/
func (rn *raft) ApplyMsg(ctx context.Context, pCommand Command) error {
	insertLog := oneRaftLog{command: pCommand, commandSerial: pCommand.ToString()}
	rn.stateMutex.Lock()
	insertLog.term = rn.currentTerm
	if rn.currentTerm != leader {
		rn.stateMutex.Unlock()
		return fmt.Errorf("Node Not Leader")
	}
	msgIndex, prevIndex, prevTerm, err := rn.logs.append(insertLog, rn.currentTerm)
	if err != nil {
		rn.stateMutex.Unlock()
		return err
	}
	rn.stateMutex.Unlock()
	//等待ctx超时
	syncResultChannel := make(chan bool)
	//同步数据
	var successNum int32
	for key, value := range rn.otherClusterNodes {
		//赋值给临时变量，否则，可能协程中引用同一个变量
		tmpkey := key
		tmpValue := value
		go func() {
			resp, err := tmpValue.GetRPC().AppendEntires(ctx, &rpc.AppendEntiresReq{LeaderTerm: insertLog.term, LeaderId: rn.localNodeMeta.NodeId, PrevLogIndex: prevIndex, PrevLogTerm: prevTerm})
			if err != nil || resp == nil {
				rn.otherNodeRetryChannel[tmpkey] <- retryAsyncLogWrapper{index: msgIndex, termNotMatch: false, respChannel: syncResultChannel}
				return
			}
			//如果发现一个Term比当前节点大，直接变为follower就可以，这里不会跟commit的逻辑冲突，直接认为自己失败
			if resp.Term > insertLog.term {
				//更新节点为follower
				rn.stateMutex.Lock()
				if resp.Term > rn.currentTerm {
					rn.currentTerm = resp.Term
					rn.state = follower
					rn.voteFor = ""
					rn.setNextRetryVote()
				}
				rn.stateMutex.Unlock()
				return
			}
			if !resp.Success {
				rn.otherNodeRetryChannel[tmpkey] <- retryAsyncLogWrapper{index: msgIndex, termNotMatch: true, respChannel: syncResultChannel}
				return
			}
			atomic.AddInt32(&successNum, 1)
			if atomic.LoadInt32(&successNum)+1 == rn.clusterMajority {
				rn.logs.commit(msgIndex)
				syncResultChannel <- true
			}
			//更新nextIndex和matchIndex
			rn.matchAndNextIndexMutex.Lock()
			if rn.nextIndex[tmpkey] <= msgIndex {
				rn.nextIndex[tmpkey] = msgIndex + 1
			}
			if rn.matchIndex[tmpkey] < msgIndex {
				rn.matchIndex[tmpkey] = msgIndex
			}
			rn.matchAndNextIndexMutex.Unlock()
		}()
	}
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
func (rn *raft) asyncLogRoutine(pNodeKey int) {
	for {
		select {
		case <-rn.cancelContext.Done():
			return
		case retryWrapper := <-rn.otherNodeRetryChannel[pNodeKey]:
			rn.matchAndNextIndexMutex.RLock()
			nextIndex := rn.nextIndex[pNodeKey]
			rn.matchAndNextIndexMutex.RUnlock()
			if nextIndex > retryWrapper.index {
				//在之前的重试中，已经完成了同步
				break
			}
			rn.stateMutex.RLock()
			if rn.state != leader {
				rn.stateMutex.RUnlock()
				break
			}
			tmpCurrentTerm := rn.currentTerm
			// 不可能返回EIndexGreaterThanMax
			// TODO 判断如果返回EIndexLessThanMin，则同步快照
			prevTerm, logs, afterEndIndex, _ := rn.logs.get(nextIndex, rn.config.retryBatchSyncLogNum)
			rn.stateMutex.RUnlock()
			for {
				ctx, _ := context.WithTimeout(context.Background(), time.Duration(rn.config.rpcTimeoutTimeMs)*time.Millisecond)
				resp, err := rn.otherClusterNodes[pNodeKey].GetRPC().AppendEntires(ctx, &rpc.AppendEntiresReq{LeaderTerm: tmpCurrentTerm, LeaderId: rn.localNodeMeta.NodeId, PrevLogIndex: nextIndex - 1, PrevLogTerm: prevTerm, Entries: logs})
				if err != nil || resp == nil {
					time.Sleep(time.Duration(rn.config.retryIntervalMs))
					continue
				}
				if !resp.Success {
					// 更新nextIndex进行retry
					// 不可能超出
					// TODO 判断如果小于，则同步快照
					rn.stateMutex.RLock()
					nextTryIndex, _ := rn.logs.getNextTryWhenAppendEntiresFalse(nextIndex)
					rn.stateMutex.RUnlock()
					rn.matchAndNextIndexMutex.Lock()
					rn.nextIndex[pNodeKey] = nextTryIndex
					rn.matchAndNextIndexMutex.Unlock()
					continue
				}

				lastLogIndex, _, _ := rn.logs.back()
				rn.matchAndNextIndexMutex.Lock()
				rn.nextIndex[pNodeKey] = afterEndIndex
				rn.matchIndex[pNodeKey] = afterEndIndex - 1
				matchIndexSlice := make([]uint64, len(rn.matchIndex)+1)
				copy(matchIndexSlice, rn.matchIndex)
				matchIndexSlice[len(rn.matchIndex)] = lastLogIndex
				rn.matchAndNextIndexMutex.Unlock()

				//尝试检查是否需要commit
				sort.Slice(matchIndexSlice, func(i, j int) bool { return matchIndexSlice[i] > matchIndexSlice[j] })
				//本次同步的日志的交集，如果是第majority，则尝试同步
				beginCommitIndex := nextIndex
				endCommitIndex := afterEndIndex - 1

				if int(rn.clusterMajority+1) < len(matchIndexSlice) && matchIndexSlice[rn.clusterMajority+1] > beginCommitIndex {
					beginCommitIndex = matchIndexSlice[rn.clusterMajority+1]
				}
				if matchIndexSlice[rn.clusterMajority] < endCommitIndex {
					endCommitIndex = matchIndexSlice[rn.clusterMajority]
				}
				//依次提交
				for beginCommitIndex <= endCommitIndex {
					rn.logs.commit(beginCommitIndex)
				}

			}
		}
	}
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

/*
	! 这里不需要通过加读锁，来优化性能，因为有两种场景，一种是心跳，50ms一次，一种是拷贝日志，需要对日志对象加写锁
*/
func (rn *raft) AppendEntires(ctx context.Context, req *rpc.AppendEntiresReq) (*rpc.AppendEntiresResp, error) {
	oneRaftLogSlice := make([]oneRaftLog, len(req.Entries))
	for key, value := range req.Entries {
		oneRaftLogSlice[key] = oneRaftLog{}
	}
	rn.stateMutex.Lock()
	defer rn.stateMutex.Unlock()
	if req.LeaderTerm < rn.currentTerm {
		return &rpc.AppendEntiresResp{Term: rn.currentTerm, Success: false}, nil
	}
	if req.LeaderTerm > rn.currentTerm {
		rn.currentTerm = req.LeaderTerm
		rn.state = follower
		rn.voteFor = ""
	} else if rn.state != follower {
		rn.state = candidate
	}
	rn.setNextRetryVote()
	//做insert操作
	err := rn.logs.insert(req.PrevLogTerm, req.PrevLogIndex, req.Entries, req.LeaderTerm)
	if err != nil {
		if err != EIndexGreaterThanMax {
			return &rpc.AppendEntiresResp{Term: req.LeaderTerm, Success: false}, err
		}
		respChannel := make(chan error)
		heap.Push(rn.waitInsertQueue, appendEntiresReqWithRespChannel{req, respChannel})
		rn.stateMutex.Unlock()
		select {
		case <-ctx.Done():
			return &rpc.AppendEntiresResp{Term: req.LeaderTerm, Success: false}, ETimeOut
		case resp := <-respChannel:
			if resp == nil {
				return &rpc.AppendEntiresResp{Term: req.LeaderTerm, Success: true}, nil
			}
			return &rpc.AppendEntiresResp{Term: req.LeaderTerm, Success: false}, resp
		}
	}

	//TODO 先返回再尝试处理其他请求
	//尝试从channel中获取数据，然后做insert操作
	for rn.waitInsertQueue.Len() > 0 {
		top := rn.waitInsertQueue.Top().(*appendEntiresReqWithRespChannel)
		err := rn.logs.insert(req.PrevLogTerm, req.PrevLogIndex, req.Entries, req.LeaderTerm)
		if err != nil {
			if err != EIndexGreaterThanMax {
				top.resp <- err
				heap.Pop(rn.waitInsertQueue)
			}
			//如果存在空隙，不再尝试
			break
		}
		top.resp <- nil
		heap.Pop(rn.waitInsertQueue)
	}
	return &rpc.AppendEntiresResp{Term: rn.currentTerm, Success: true}, nil
}

func (rn *raft) RequestVote(ctx context.Context, req *rpc.RequestVoteReq) (*rpc.RequestVoteResp, error) {
	//RequestVote理论上应该是小概率事件，所以直接全局加锁
	isToDataLastLog := rn.logs.upToDateLast(req.LastLogTerm, req.LastLogIndex)
	rn.stateMutex.Lock()
	defer rn.stateMutex.Unlock()
	log.Instance().Debugf("node{id:%v, term:%v} get request vote from node{id:%v, term:%v }", rn.localNodeMeta.NodeId, rn.currentTerm, req.CandidateId, req.Term)
	if req.Term < rn.currentTerm {
		return &rpc.RequestVoteResp{Term: rn.currentTerm, VoteGranted: false}, nil
	} else if req.Term == rn.currentTerm {
		if rn.state == follower && (rn.voteFor == "" || rn.voteFor == req.CandidateId) && isToDataLastLog {
			rn.voteFor = req.CandidateId
			//刷新时钟
			rn.setNextRetryVote()
			return &rpc.RequestVoteResp{Term: rn.currentTerm, VoteGranted: true}, nil
		}
		return &rpc.RequestVoteResp{Term: rn.currentTerm, VoteGranted: false}, nil
	}
	log.Instance().Debugf("node[%v] update to term [%v] in RequestVote()", rn.localNodeMeta.NodeId, req.Term)
	rn.currentTerm = req.Term
	rn.state = follower
	if isToDataLastLog {
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
				log.Instance().Debugf("node[%v] update to term [%v] in tickHeartBeat()", rn.localNodeMeta.NodeId, maxRespTerm)
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
			log.Instance().Debugf("node[%v] begin request vote term[%v] time[%v]", rn.localNodeMeta.NodeId, tmpCurrentTerm, now.UnixMilli())
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
			log.Instance().Debugf("node[%v] vote num[%v]", rn.localNodeMeta.NodeId, voteNum)
			if voteNum >= rn.clusterMajority {
				rn.stateMutex.Lock()
				if rn.currentTerm == tmpCurrentTerm {
					//优化启停
					rn.state = leader
					log.Instance().Debugf("node[%v] become leader", rn.localNodeMeta.NodeId)
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
				log.Instance().Debugf("node[%v] RequestVote(%v) error[%v]", rn.localNodeMeta.NodeId, tmpNode.GetNodeMeta(), err)
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
		log.Instance().Debugf("node[%v] get request vote resp[%v]", rn.localNodeMeta.NodeId, *oneResp)
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
