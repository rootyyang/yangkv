package consensus

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/domain/node"
	"github.com/rootyyang/yangkv/pkg/infrastructure/log"
	"github.com/rootyyang/yangkv/pkg/infrastructure/rpc"
	"github.com/rootyyang/yangkv/pkg/infrastructure/system"
)

func CheckOneLeader(raftNodes []raft, isConnect []bool) (leaderIndex int, leaderTerm uint64) {
	time.Sleep(time.Duration(450) * time.Millisecond)
	var maxTerm uint64
	leaderIndex = -1
	term2Leader := make(map[uint64]int)
	for key, value := range raftNodes {
		if !isConnect[key] {
			continue
		}
		oneState, oneTerm := value.GetRaftState()
		if oneState == leader {
			if _, ok := term2Leader[oneTerm]; ok {
				return
			}
			term2Leader[oneTerm] = key
		}
		if oneTerm > maxTerm {
			maxTerm = oneTerm
		}

	}
	if _, ok := term2Leader[maxTerm]; ok {
		leaderIndex = term2Leader[maxTerm]
		leaderTerm = maxTerm
	}
	return
}

func GetMinAndMaxTerm(raftNodes []raft) (minTerm, maxTerm uint64) {
	minTerm = math.MaxUint64
	for _, value := range raftNodes {
		_, term := value.GetRaftState()
		if term > maxTerm {
			maxTerm = term
		}
		if term < minTerm {
			minTerm = term
		}
	}
	return
}

func CreateNodeMeta(pNodeNumber int) []node.NodeMeta {
	ret := make([]node.NodeMeta, pNodeNumber)
	for i := 0; i < pNodeNumber; i++ {
		ret[i] = node.NodeMeta{NodeId: "NodeId" + strconv.Itoa(i), Ip: "127.0.0.1", Port: "688" + strconv.Itoa(i), ClusterName: "TestClusterName", ClusterId: "TestClusterID"}
	}
	return ret
}

func CreateRaftNode(pNodeNumber int) ([]raft, []node.NodeMeta, []bool) {
	allNodeMeta := CreateNodeMeta(pNodeNumber)
	ret := make([]raft, pNodeNumber)
	isConnect := make([]bool, pNodeNumber)
	for i := 0; i < pNodeNumber; i++ {
		ret[i] = raft{}
		ret[i].Start(allNodeMeta[i], allNodeMeta)
		isConnect[i] = true
	}
	return ret, allNodeMeta, isConnect
}
func StopRaftNode(pAllRafts []raft) {
	for _, value := range pAllRafts {
		value.Stop()
	}
}

func disconnect(pRaft raft) {
	pRaft.rpcServer.Stop()
	for _, value := range pRaft.otherClusterNodes {
		value.GetRPC().Stop()
	}
	fmt.Printf("node[%v] disconnect\n", pRaft.localNodeMeta.NodeId)
}
func reConnect(pRaft raft) {
	pRaft.rpcServer.Start()
	for _, value := range pRaft.otherClusterNodes {
		value.GetRPC().Start()
	}
	time.Sleep(time.Second)
}

func CreateOneRaftWithMockRPC(pNodeNumber int, mockCtrl *gomock.Controller) (*raft, []*rpc.MockRPClient) {
	mockClientProvider, mockServerProvider := rpc.RegisterAndUseMockRPC(mockCtrl)
	allNodeMeta := CreateNodeMeta(pNodeNumber)
	retRaft := &raft{}
	retMockRPCClient := make([]*rpc.MockRPClient, pNodeNumber-1)
	mockCalls := make([]*gomock.Call, 0)
	mockServer := rpc.NewMockRPCServer(mockCtrl)
	for i := 0; i < pNodeNumber-1; i++ {
		retMockRPCClient[i] = rpc.NewMockRPClient(mockCtrl)
		oneMockCall := mockClientProvider.EXPECT().GetRPCClient(gomock.Any(), gomock.Any()).Return(retMockRPCClient[i])
		mockCalls = append(mockCalls, oneMockCall)
		retMockRPCClient[i].EXPECT().Start().Times(1)
		retMockRPCClient[i].EXPECT().Stop().Times(1)
	}
	mockServerProvider.EXPECT().GetRPCServer(gomock.Any()).Return(mockServer)
	mockServer.EXPECT().Start()
	mockServer.EXPECT().RegisterRaftHandle(gomock.Any())
	mockServer.EXPECT().Stop()
	gomock.InOrder(mockCalls...)
	retRaft.Start(allNodeMeta[0], allNodeMeta)
	time.Sleep(10 * time.Millisecond)
	return retRaft, retMockRPCClient
}

func TestOneRaftLeaderBecomeLeader(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	log.RegisterAndUseMockLog(mockCtrl)
	mockClock := system.RegisterAndUseMockClock(mockCtrl)
	mockTimer := system.NewMockClockTimer(mockCtrl)
	mockTicker := system.NewMockClockTicker(mockCtrl)
	mockRandom := system.RegisterAndUseMockRandom(mockCtrl)
	mockRandom.EXPECT().Int().Return(150).AnyTimes()

	timerChannel := make(chan time.Time)
	tickerChannel := make(chan time.Time)
	mockTimer.EXPECT().GetChannel().AnyTimes().Return(timerChannel)
	mockTimer.EXPECT().Reset(gomock.Any()).Return(true).AnyTimes()
	mockTicker.EXPECT().GetChannel().AnyTimes().Return(tickerChannel)
	mockClock.EXPECT().Ticker(gomock.Any()).Return(mockTicker)
	mockClock.EXPECT().Timer(gomock.Any()).Return(mockTimer)
	mockTicker.EXPECT().Stop()

	//阶段一，服务刚刚启动
	baseTime := time.Now()
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	nodeNumber := 7
	raft, mockRPCClients := CreateOneRaftWithMockRPC(nodeNumber, mockCtrl)
	if len(mockRPCClients) != nodeNumber-1 {
		t.Fatalf("len(mockRPCClients)=%v, want = %v", len(mockRPCClients), nodeNumber-1)
	}

	//阶段二, 节点选为leader
	baseTime = baseTime.Add(time.Duration(351) * time.Millisecond)
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	requestVoteResp := rpc.RequestVoteResp{1, true}
	for i := 0; i < 2; i++ {
		mockRPCClients[i].EXPECT().RequestVote(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *rpc.RequestVoteReq) (*rpc.RequestVoteResp, error) {
			if req.Term != 1 || req.CandidateId != "NodeId0" {
				t.Fatalf("RequestVoteReq=%v, want = {Term:1, CandidateId:NodeId0}", req)
			}
			return &requestVoteResp, nil
		})
	}
	requestVoteRespReject := rpc.RequestVoteResp{1, false}
	for i := 2; i < nodeNumber-1; i++ {
		mockRPCClients[i].EXPECT().RequestVote(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *rpc.RequestVoteReq) (*rpc.RequestVoteResp, error) {
			if req.Term != 1 || req.CandidateId != "NodeId0" {
				t.Fatalf("RequestVoteReq=%v, want = {Term:1, CandidateId:NodeId0}", req)
			}
			return &requestVoteRespReject, nil
		})
	}

	timerChannel <- baseTime
	time.Sleep(10 * time.Millisecond)
	state, term := raft.GetRaftState()
	if state == leader {
		t.Fatalf("raft state=[%v] want != leader", state)
	}
	if term != 1 {
		t.Fatalf("raft term=[%v]  want=1", term)
	}
	raft.Stop()
}

func TestOneRaftLeaderAppend(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	log.RegisterAndUseMockLog(mockCtrl)
	mockClock := system.RegisterAndUseMockClock(mockCtrl)
	mockTimer := system.NewMockClockTimer(mockCtrl)
	mockTicker := system.NewMockClockTicker(mockCtrl)
	mockRandom := system.RegisterAndUseMockRandom(mockCtrl)
	mockRandom.EXPECT().Int().Return(150).AnyTimes()

	timerChannel := make(chan time.Time)
	tickerChannel := make(chan time.Time)
	mockTimer.EXPECT().GetChannel().AnyTimes().Return(timerChannel)
	mockTimer.EXPECT().Reset(gomock.Any()).Return(true).AnyTimes()
	mockTicker.EXPECT().GetChannel().AnyTimes().Return(tickerChannel)
	mockClock.EXPECT().Ticker(gomock.Any()).Return(mockTicker)
	mockClock.EXPECT().Timer(gomock.Any()).Return(mockTimer)
	mockTicker.EXPECT().Stop()

	//阶段一，服务刚刚启动
	baseTime := time.Now()
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	nodeNumber := 3
	raft, mockRPCClients := CreateOneRaftWithMockRPC(nodeNumber, mockCtrl)
	if len(mockRPCClients) != nodeNumber-1 {
		t.Fatalf("len(mockRPCClients)=%v, want = %v", len(mockRPCClients), nodeNumber-1)
	}

	//阶段二, 节点选为leader
	baseTime = baseTime.Add(time.Duration(351) * time.Millisecond)
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	requestVoteResp := rpc.RequestVoteResp{1, true}
	for i := 0; i < len(mockRPCClients); i++ {
		mockRPCClients[i].EXPECT().RequestVote(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *rpc.RequestVoteReq) (*rpc.RequestVoteResp, error) {
			if req.Term != 1 || req.CandidateId != "NodeId0" {
				t.Fatalf("RequestVoteReq=%v, want = {Term:1, CandidateId:NodeId0}", req)
			}
			return &requestVoteResp, nil
		})
	}

	timerChannel <- baseTime
	time.Sleep(10 * time.Millisecond)
	state, term := raft.GetRaftState()
	if state != leader {
		t.Fatalf("raft state=[%v] want=leader", state)
	}
	if term != 1 {
		t.Fatalf("raft term=[%v]  want=1", term)
	}
	//阶段三，接收到Term = 0 的 Append
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	appendReq := &rpc.AppendEntiresReq{LeaderTerm: 0, LeaderId: "NodeID2"}
	appendResp, err := raft.AppendEntires(ctx, appendReq)
	if appendResp == nil || err != nil {
		t.Fatalf("raft.AppendEntires()=[%v][%v]  want not nil, nil", appendResp, err)
	}
	if appendResp.Success == true || appendResp.Term != 1 {
		t.Fatalf("raft.AppendEntires()=[%v]  want {false, 1}", *appendResp)
	}
	state, term = raft.GetRaftState()
	if state != leader {
		t.Fatalf("raft state=[%v] want=leader", state)
	}
	if term != 1 {
		t.Fatalf("raft term=[%v]  want=1", term)
	}
	//阶段四，接收到Term = 2 的Append
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	appendReq = &rpc.AppendEntiresReq{LeaderTerm: 2, LeaderId: "NodeID2"}
	appendResp, err = raft.AppendEntires(ctx, appendReq)
	if appendResp == nil || err != nil {
		t.Fatalf("raft.AppendEntires()=[%v][%v]  want not nil, nil", appendResp, err)
	}
	if appendResp.Success == false || appendResp.Term != 2 {
		t.Fatalf("raft.AppendEntires()=[%v]  want {false, 1}", *appendResp)
	}
	state, term = raft.GetRaftState()
	if state != follower {
		t.Fatalf("raft state=[%v] want=follower", state)
	}
	if term != 2 {
		t.Fatalf("raft term=[%v]  want=2", term)
	}
	raft.Stop()
}

func TestOneRaftLeaderRequestVote(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	log.RegisterAndUseMockLog(mockCtrl)
	mockClock := system.RegisterAndUseMockClock(mockCtrl)
	mockTimer := system.NewMockClockTimer(mockCtrl)
	mockTicker := system.NewMockClockTicker(mockCtrl)
	mockRandom := system.RegisterAndUseMockRandom(mockCtrl)
	mockRandom.EXPECT().Int().Return(150).AnyTimes()

	timerChannel := make(chan time.Time)
	tickerChannel := make(chan time.Time)
	mockTimer.EXPECT().GetChannel().AnyTimes().Return(timerChannel)
	mockTimer.EXPECT().Reset(gomock.Any()).Return(true).AnyTimes()
	mockTicker.EXPECT().GetChannel().AnyTimes().Return(tickerChannel)
	mockClock.EXPECT().Ticker(gomock.Any()).Return(mockTicker)
	mockClock.EXPECT().Timer(gomock.Any()).Return(mockTimer)
	mockTicker.EXPECT().Stop()
	//阶段一，服务刚刚启动
	baseTime := time.Now()
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	nodeNumber := 3
	raft, mockRPCClients := CreateOneRaftWithMockRPC(nodeNumber, mockCtrl)
	if len(mockRPCClients) != nodeNumber-1 {
		t.Fatalf("len(mockRPCClients)=%v, want = %v", len(mockRPCClients), nodeNumber-1)
	}

	//阶段二, 节点选为leader
	baseTime = baseTime.Add(time.Duration(351) * time.Millisecond)
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	requestVoteResp := rpc.RequestVoteResp{1, true}
	for i := 0; i < len(mockRPCClients); i++ {
		mockRPCClients[i].EXPECT().RequestVote(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *rpc.RequestVoteReq) (*rpc.RequestVoteResp, error) {
			if req.Term != 1 || req.CandidateId != "NodeId0" {
				t.Fatalf("RequestVoteReq=%v, want = {Term:1, CandidateId:NodeId0}", req)
			}
			return &requestVoteResp, nil
		})
	}

	timerChannel <- baseTime
	time.Sleep(10 * time.Millisecond)
	state, term := raft.GetRaftState()
	if state != leader {
		t.Fatalf("raft state=[%v] want=leader", state)
	}
	if term != 1 {
		t.Fatalf("raft term=[%v]  want=1", term)
	}

	//阶段三 接收到node1的选举请求，Term为0，拒接该请求
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	nodeRequestVote := rpc.RequestVoteReq{Term: 0, CandidateId: "NodeID1"}
	nodeRequestVoteResp, err := raft.RequestVote(ctx, &nodeRequestVote)
	if err != nil || nodeRequestVoteResp == nil {
		t.Fatalf("raft.RequestVote()=[%v][%v]  want nil, not nil", nodeRequestVoteResp, err)
	}
	if nodeRequestVoteResp.Term != 1 || nodeRequestVoteResp.VoteGranted != false {
		t.Fatalf("nodeRequestVote=[%v]  want {1, false}", *nodeRequestVoteResp)
	}

	state, term = raft.GetRaftState()
	if state != leader {
		t.Fatalf("raft state=[%v] want=leader", state)
	}
	if term != 1 {
		t.Fatalf("raft term=[%v]  want=1", term)
	}

	//阶段四 接收到node1的选举请求，Term为1，接受该请求
	nodeRequestVote = rpc.RequestVoteReq{Term: 1, CandidateId: "NodeID1"}
	nodeRequestVoteResp, err = raft.RequestVote(ctx, &nodeRequestVote)
	if err != nil || nodeRequestVoteResp == nil {
		t.Fatalf("raft.RequestVote()=[%v][%v]  want nil, not nil", nodeRequestVoteResp, err)
	}
	if nodeRequestVoteResp.Term != 1 || nodeRequestVoteResp.VoteGranted != false {
		t.Fatalf("nodeRequestVote=[%v]  want {1, false}", *nodeRequestVoteResp)
	}
	state, term = raft.GetRaftState()
	if state != leader {
		t.Fatalf("raft state=[%v] want=leader", state)
	}
	if term != 1 {
		t.Fatalf("raft term=[%v]  want=1", term)
	}

	//阶段五， 接收到node2的选举请求，Term为2，拒绝该请求
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	nodeRequestVote = rpc.RequestVoteReq{Term: 2, CandidateId: "NodeID2"}
	nodeRequestVoteResp, err = raft.RequestVote(ctx, &nodeRequestVote)
	if err != nil || nodeRequestVoteResp == nil {
		t.Fatalf("raft.RequestVote()=[%v][%v]  want nil, not nil", nodeRequestVoteResp, err)
	}
	if nodeRequestVoteResp.Term != 2 || nodeRequestVoteResp.VoteGranted != true {
		t.Fatalf("nodeRequestVote=[%v]  want {2, false}", *nodeRequestVoteResp)
	}

	state, term = raft.GetRaftState()
	if state != follower {
		t.Fatalf("raft state=[%v] want=follower", state)
	}
	if term != 2 {
		t.Fatalf("raft term=[%v]  want=2", term)
	}
	raft.Stop()
}

func TestOneRaftFollowerRequestVote(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	log.RegisterAndUseMockLog(mockCtrl)
	mockClock := system.RegisterAndUseMockClock(mockCtrl)
	mockTimer := system.NewMockClockTimer(mockCtrl)
	mockTicker := system.NewMockClockTicker(mockCtrl)
	mockRandom := system.RegisterAndUseMockRandom(mockCtrl)
	mockRandom.EXPECT().Int().Return(150).AnyTimes()

	timerChannel := make(chan time.Time)
	tickerChannel := make(chan time.Time)
	mockTimer.EXPECT().GetChannel().AnyTimes().Return(timerChannel)
	mockTimer.EXPECT().Reset(gomock.Any()).Return(true).AnyTimes()
	mockTicker.EXPECT().GetChannel().AnyTimes().Return(tickerChannel)
	mockClock.EXPECT().Ticker(gomock.Any()).Return(mockTicker)
	mockClock.EXPECT().Timer(gomock.Any()).Return(mockTimer)
	mockTicker.EXPECT().Stop()
	//阶段一，服务刚刚启动
	baseTime := time.Now()
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	nodeNumber := 3
	raft, mockRPCClients := CreateOneRaftWithMockRPC(nodeNumber, mockCtrl)
	if len(mockRPCClients) != nodeNumber-1 {
		t.Fatalf("len(mockRPCClients)=%v, want = %v", len(mockRPCClients), nodeNumber-1)
	}

	//阶段二，raft接收到RequestVote Term = 1 Id = node1
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	nodeRequestVote := rpc.RequestVoteReq{Term: 1, CandidateId: "NodeID1"}
	nodeRequestVoteResp, err := raft.RequestVote(ctx, &nodeRequestVote)
	if err != nil || nodeRequestVoteResp == nil {
		t.Fatalf("raft.RequestVote()=[%v][%v]  want nil, not nil", nodeRequestVoteResp, err)
	}
	if nodeRequestVoteResp.Term != 1 || nodeRequestVoteResp.VoteGranted != true {
		t.Fatalf("nodeRequestVote=[%v]  want {1, true}", *nodeRequestVoteResp)
	}

	state, term := raft.GetRaftState()
	if state != follower {
		t.Fatalf("raft state=[%v] want=follower", state)
	}
	if term != 1 {
		t.Fatalf("raft term=[%v]  want=2", term)
	}
	//阶段三，raft接收到RequestVote Term = 1 Id = node2
	nodeRequestVote = rpc.RequestVoteReq{Term: 1, CandidateId: "NodeID2"}
	nodeRequestVoteResp, err = raft.RequestVote(ctx, &nodeRequestVote)
	if err != nil || nodeRequestVoteResp == nil {
		t.Fatalf("raft.RequestVote()=[%v][%v]  want nil, not nil", nodeRequestVoteResp, err)
	}
	if nodeRequestVoteResp.Term != 1 || nodeRequestVoteResp.VoteGranted != false {
		t.Fatalf("nodeRequestVote=[%v]  want {1, false}", *nodeRequestVoteResp)
	}

	state, term = raft.GetRaftState()
	if state != follower {
		t.Fatalf("raft state=[%v] want=follower", state)
	}
	if term != 1 {
		t.Fatalf("raft term=[%v]  want=2", term)
	}

	//阶段四，raft接收到RequestVote Term = 2 Id = node3
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	nodeRequestVote = rpc.RequestVoteReq{Term: 2, CandidateId: "NodeID3"}
	nodeRequestVoteResp, err = raft.RequestVote(ctx, &nodeRequestVote)
	if err != nil || nodeRequestVoteResp == nil {
		t.Fatalf("raft.RequestVote()=[%v][%v]  want nil, not nil", nodeRequestVoteResp, err)
	}
	if nodeRequestVoteResp.Term != 2 || nodeRequestVoteResp.VoteGranted != true {
		t.Fatalf("nodeRequestVote=[%v]  want {2, true}", *nodeRequestVoteResp)
	}

	state, term = raft.GetRaftState()
	if state != follower {
		t.Fatalf("raft state=[%v] want=follower", state)
	}
	if term != 2 {
		t.Fatalf("raft term=[%v]  want=2", term)
	}
	raft.Stop()
}

func TestOneRaftFollowerAppendEntries(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	log.RegisterAndUseMockLog(mockCtrl)
	mockClock := system.RegisterAndUseMockClock(mockCtrl)
	mockTimer := system.NewMockClockTimer(mockCtrl)
	mockTicker := system.NewMockClockTicker(mockCtrl)
	mockRandom := system.RegisterAndUseMockRandom(mockCtrl)
	mockRandom.EXPECT().Int().Return(150).AnyTimes()

	timerChannel := make(chan time.Time)
	tickerChannel := make(chan time.Time)
	mockTimer.EXPECT().GetChannel().AnyTimes().Return(timerChannel)
	mockTimer.EXPECT().Reset(gomock.Any()).Return(true).AnyTimes()
	mockTicker.EXPECT().GetChannel().AnyTimes().Return(tickerChannel)
	mockClock.EXPECT().Ticker(gomock.Any()).Return(mockTicker)
	mockClock.EXPECT().Timer(gomock.Any()).Return(mockTimer)
	mockTicker.EXPECT().Stop()
	//阶段一，服务刚刚启动
	baseTime := time.Now()
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	nodeNumber := 3
	raft, mockRPCClients := CreateOneRaftWithMockRPC(nodeNumber, mockCtrl)
	if len(mockRPCClients) != nodeNumber-1 {
		t.Fatalf("len(mockRPCClients)=%v, want = %v", len(mockRPCClients), nodeNumber-1)
	}

	//阶段二，接收到Term = 1 的 Append
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	appendReq := &rpc.AppendEntiresReq{LeaderTerm: 1, LeaderId: "NodeID1"}
	appendResp, err := raft.AppendEntires(ctx, appendReq)
	if appendResp == nil || err != nil {
		t.Fatalf("raft.AppendEntires()=[%v][%v]  want not nil, nil", appendResp, err)
	}
	if appendResp.Success != true || appendResp.Term != 1 {
		t.Fatalf("raft.AppendEntires()=[%v]  want {false, 1}", *appendResp)
	}
	state, term := raft.GetRaftState()
	if state != follower {
		t.Fatalf("raft state=[%v] want=leader", state)
	}
	if term != 1 {
		t.Fatalf("raft term=[%v]  want=1", term)
	}

	//阶段四，接收到Term = 0 的 Append
	appendReq = &rpc.AppendEntiresReq{LeaderTerm: 0, LeaderId: "NodeID2"}
	appendResp, err = raft.AppendEntires(ctx, appendReq)
	if appendResp == nil || err != nil {
		t.Fatalf("raft.AppendEntires()=[%v][%v]  want not nil, nil", appendResp, err)
	}
	if appendResp.Success != false || appendResp.Term != 1 {
		t.Fatalf("raft.AppendEntires()=[%v]  want {false, 1}", *appendResp)
	}
	state, term = raft.GetRaftState()
	if state != follower {
		t.Fatalf("raft state=[%v] want=follower", state)
	}
	if term != 1 {
		t.Fatalf("raft term=[%v]  want=2", term)
	}

	//阶段五，接收到 Term = 1 的Append
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	appendReq = &rpc.AppendEntiresReq{LeaderTerm: 1, LeaderId: "NodeID3"}
	appendResp, err = raft.AppendEntires(ctx, appendReq)
	if appendResp == nil || err != nil {
		t.Fatalf("raft.AppendEntires()=[%v][%v]  want not nil, nil", appendResp, err)
	}
	if appendResp.Success != true || appendResp.Term != 1 {
		t.Fatalf("raft.AppendEntires()=[%v]  want {true, 1}", *appendResp)
	}
	state, term = raft.GetRaftState()
	if state != follower {
		t.Fatalf("raft state=[%v] want=follower", state)
	}
	if term != 1 {
		t.Fatalf("raft term=[%v]  want=2", term)
	}

	//阶段六，接收到 Term = 1 的Append
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	appendReq = &rpc.AppendEntiresReq{LeaderTerm: 2, LeaderId: "NodeID3"}
	appendResp, err = raft.AppendEntires(ctx, appendReq)
	if appendResp == nil || err != nil {
		t.Fatalf("raft.AppendEntires()=[%v][%v]  want not nil, nil", appendResp, err)
	}
	if appendResp.Success != true || appendResp.Term != 2 {
		t.Fatalf("raft.AppendEntires()=[%v]  want {true, 2}", *appendResp)
	}
	state, term = raft.GetRaftState()
	if state != follower {
		t.Fatalf("raft state=[%v] want=follower", state)
	}
	if term != 2 {
		t.Fatalf("raft term=[%v]  want=2", term)
	}
	raft.Stop()
}

func TestOneRaftBecomeLeader(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	log.RegisterAndUseMockLog(mockCtrl)
	mockClock := system.RegisterAndUseMockClock(mockCtrl)
	mockTimer := system.NewMockClockTimer(mockCtrl)
	mockTicker := system.NewMockClockTicker(mockCtrl)

	mockRandom := system.RegisterAndUseMockRandom(mockCtrl)
	mockRandom.EXPECT().Int().Return(150).AnyTimes()

	timerChannel := make(chan time.Time)
	tickerChannel := make(chan time.Time)
	mockTimer.EXPECT().GetChannel().AnyTimes().Return(timerChannel)
	mockTimer.EXPECT().Reset(gomock.Any()).Return(true).AnyTimes()
	mockTicker.EXPECT().GetChannel().AnyTimes().Return(tickerChannel)
	mockClock.EXPECT().Ticker(gomock.Any()).Return(mockTicker)
	mockClock.EXPECT().Timer(gomock.Any()).Return(mockTimer)
	mockTicker.EXPECT().Stop()

	//阶段一，服务刚刚启动
	baseTime := time.Now()
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	nodeNumber := 3
	raft, mockRPCClients := CreateOneRaftWithMockRPC(nodeNumber, mockCtrl)
	if len(mockRPCClients) != nodeNumber-1 {
		t.Fatalf("len(mockRPCClients)=%v, want = %v", len(mockRPCClients), nodeNumber-1)
	}

	//阶段二, 节点选为leader
	baseTime = baseTime.Add(time.Duration(351) * time.Millisecond)
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	requestVoteResp := rpc.RequestVoteResp{1, true}
	for i := 0; i < len(mockRPCClients); i++ {
		mockRPCClients[i].EXPECT().RequestVote(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *rpc.RequestVoteReq) (*rpc.RequestVoteResp, error) {
			if req.Term != 1 || req.CandidateId != "NodeId0" {
				t.Fatalf("RequestVoteReq=%v, want = {Term:1, CandidateId:NodeId0}", req)
			}
			return &requestVoteResp, nil
		})
	}

	timerChannel <- baseTime
	time.Sleep(10 * time.Millisecond)
	state, term := raft.GetRaftState()
	if state != leader {
		t.Fatalf("raft state=[%v] want=leader", state)
	}
	if term != 1 {
		t.Fatalf("raft term=[%v]  want=1", term)
	}

	//阶段三, leader发送心跳
	baseTime = baseTime.Add(time.Duration(351) * time.Millisecond)
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	appendEntriesResp := rpc.AppendEntiresResp{1, true}
	for i := 0; i < len(mockRPCClients); i++ {
		mockRPCClients[i].EXPECT().AppendEntires(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *rpc.AppendEntiresReq) (*rpc.AppendEntiresResp, error) {
			if req.LeaderTerm != 1 || req.LeaderId != "NodeId0" {
				t.Fatalf("RequestVoteReq=%v, want = {Term:1, CandidateId:NodeId0}", req)
			}
			return &appendEntriesResp, nil
		})
	}
	tickerChannel <- baseTime
	time.Sleep(10 * time.Millisecond)
	state, term = raft.GetRaftState()
	if state != leader {
		t.Fatalf("raft state=[%v] want=leader", state)
	}
	if term != 1 {
		t.Fatalf("raft term=[%v]  want=1", term)
	}
	//阶段四: leader发送心跳，发送版本号更高的节点
	baseTime = baseTime.Add(time.Duration(351) * time.Millisecond)
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	newAppendEntriesResp := rpc.AppendEntiresResp{2, false}
	mockRPCClients[0].EXPECT().AppendEntires(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *rpc.AppendEntiresReq) (*rpc.AppendEntiresResp, error) {
		if req.LeaderTerm != 1 || req.LeaderId != "NodeId0" {
			t.Fatalf("RequestVoteReq=%v, want = {Term:1, CandidateId:NodeId0}", req)
		}
		return &newAppendEntriesResp, nil
	})
	mockRPCClients[1].EXPECT().AppendEntires(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *rpc.AppendEntiresReq) (*rpc.AppendEntiresResp, error) {
		if req.LeaderTerm != 1 || req.LeaderId != "NodeId0" {
			t.Fatalf("RequestVoteReq=%v, want = {Term:1, CandidateId:NodeId0}", req)
		}
		return &appendEntriesResp, nil
	})

	tickerChannel <- baseTime
	time.Sleep(10 * time.Millisecond)
	state, term = raft.GetRaftState()
	if state != follower {
		t.Fatalf("raft state=[%v] want=follower", state)
	}
	if term != 2 {
		t.Fatalf("raft term=[%v]  want=2", term)
	}
	//阶段五 接收到node1的选举请求，Term为1，拒接该请求
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	nodeRequestVote := rpc.RequestVoteReq{Term: 1, CandidateId: "NodeID1"}
	nodeRequestVoteResp, err := raft.RequestVote(ctx, &nodeRequestVote)
	if err != nil || nodeRequestVoteResp == nil {
		t.Fatalf("raft.RequestVote()=[%v][%v]  want nil, not nil", nodeRequestVoteResp, err)
	}
	if nodeRequestVoteResp.Term != 2 || nodeRequestVoteResp.VoteGranted != false {
		t.Fatalf("nodeRequestVote=[%v]  want {2, false}", *nodeRequestVoteResp)
	}

	state, term = raft.GetRaftState()
	if state != follower {
		t.Fatalf("raft state=[%v] want=follower", state)
	}
	if term != 2 {
		t.Fatalf("raft term=[%v]  want=2", term)
	}
	//阶段六 接收到node1的选举请求，Term为2，接受该请求
	nodeRequestVote = rpc.RequestVoteReq{Term: 2, CandidateId: "NodeID1"}
	nodeRequestVoteResp, err = raft.RequestVote(ctx, &nodeRequestVote)
	if err != nil || nodeRequestVoteResp == nil {
		t.Fatalf("raft.RequestVote()=[%v][%v]  want nil, not nil", nodeRequestVoteResp, err)
	}
	if nodeRequestVoteResp.Term != 2 || nodeRequestVoteResp.VoteGranted != true {
		t.Fatalf("nodeRequestVote=[%v]  want {2, true}", *nodeRequestVoteResp)
	}
	fmt.Printf("6 raft [%v], now [%v]\n", raft.tryRequestVoteTime, baseTime)
	//阶段七， 接收到node2的选举请求，Term为2，拒绝该请求
	nodeRequestVote = rpc.RequestVoteReq{Term: 2, CandidateId: "NodeID2"}
	nodeRequestVoteResp, err = raft.RequestVote(ctx, &nodeRequestVote)
	if err != nil || nodeRequestVoteResp == nil {
		t.Fatalf("raft.RequestVote()=[%v][%v]  want nil, not nil", nodeRequestVoteResp, err)
	}
	if nodeRequestVoteResp.Term != 2 || nodeRequestVoteResp.VoteGranted != false {
		t.Fatalf("nodeRequestVote=[%v]  want {2, false}", *nodeRequestVoteResp)
	}
	//阶段八，时间不变，不发起重新选举，只是重新sleep
	timerChannel <- baseTime
	time.Sleep(10 * time.Millisecond)

	state, term = raft.GetRaftState()
	if state != follower {
		t.Fatalf("raft state=[%v] want=follower", state)
	}
	if term != 2 {
		t.Fatalf("raft term=[%v]  want=1", term)
	}

	//阶段九，发起重新选举，重新称为leader
	baseTime = baseTime.Add(time.Duration(351) * time.Millisecond)
	mockClock.EXPECT().Now().Return(baseTime).Times(1)
	requestVoteResp = rpc.RequestVoteResp{2, true}
	for i := 0; i < len(mockRPCClients); i++ {
		mockRPCClients[i].EXPECT().RequestVote(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *rpc.RequestVoteReq) (*rpc.RequestVoteResp, error) {
			if req.Term != 3 || req.CandidateId != "NodeId0" {
				t.Fatalf("RequestVoteReq=%v, want = {Term:2, CandidateId:NodeId0}", req)
			}
			return &requestVoteResp, nil
		})
	}
	timerChannel <- baseTime
	time.Sleep(10 * time.Millisecond)
	state, term = raft.GetRaftState()
	if state != leader {
		t.Fatalf("raft state=[%v] want=leader", state)
	}
	if term != 3 {
		t.Fatalf("raft term=[%v]  want=3", term)
	}
	raft.Stop()

}

func TestMultiRaftSelectOneLeader(t *testing.T) {
	rpc.SetUseRPC("grpc")
	system.RegisterAndUseDefaultClock()
	system.RegisterAndUseDefaultRandom()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	log.RegisterAndUseMockLog(mockCtrl)

	allRaft, _, isConnect := CreateRaftNode(3)
	leaderIndex, _ := CheckOneLeader(allRaft, isConnect)
	if leaderIndex == -1 {
		t.Fatalf("Want Only One leader")
	}

	minTerm, maxTerm := GetMinAndMaxTerm(allRaft)
	if minTerm != maxTerm {
		t.Fatalf("minTerm[%v]  != maxTerm[%v], want equal", minTerm, maxTerm)
	}
	leaderIndex, _ = CheckOneLeader(allRaft, isConnect)
	if leaderIndex == -1 {
		t.Fatalf("Want Only One leader")
	}
	newMinTerm, newMaxTerm := GetMinAndMaxTerm(allRaft)
	if minTerm != newMinTerm || maxTerm != newMaxTerm {
		t.Fatalf("maxTerm[%v]  != newMaxTerm[%v], want equal", maxTerm, newMaxTerm)
	}
	StopRaftNode(allRaft)
}

func TestMultiRaftDisconnect(t *testing.T) {
	rpc.SetUseRPC("grpc")
	system.RegisterAndUseDefaultClock()
	system.RegisterAndUseDefaultRandom()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	log.RegisterAndUseMockLog(mockCtrl)
	allRaft, _, isConnect := CreateRaftNode(3)
	firstLeaderIndex, firstLeaderTerm := CheckOneLeader(allRaft, isConnect)
	if firstLeaderIndex == -1 {
		t.Fatalf("Want Only One leader")
	}
	minTerm, maxTerm := GetMinAndMaxTerm(allRaft)
	if minTerm != maxTerm {
		t.Fatalf("minTerm[%v]  != maxTerm[%v], want equal", minTerm, maxTerm)
	}
	disconnect(allRaft[firstLeaderIndex])
	secondLeaderIndex, secondLeaderTerm := CheckOneLeader(allRaft, isConnect)
	if secondLeaderIndex == -1 {
		t.Fatalf("Want Only One leader")
	}
	if firstLeaderTerm >= secondLeaderTerm {
		t.Fatalf("firstLeaderTerm[%v] >= secondLeaderTerm[%v] want <", firstLeaderTerm, secondLeaderTerm)
	}
	reConnect(allRaft[firstLeaderIndex])
	thirdLeaderIndex, thirdLeaderTerm := CheckOneLeader(allRaft, isConnect)
	if thirdLeaderIndex == -1 {
		t.Fatalf("Want Only One leader")
	}
	if thirdLeaderTerm < secondLeaderTerm {
		t.Fatalf("thirdLeaderTerm[%v] < secondLeaderTerm[%v] want >=", thirdLeaderTerm, secondLeaderTerm)
	}

	minTerm, maxTerm = GetMinAndMaxTerm(allRaft)
	if minTerm != maxTerm {
		t.Fatalf("minTerm[%v]  != maxTerm[%v], want equal", minTerm, maxTerm)
	}

	disconnect(allRaft[thirdLeaderIndex])
	isConnect[thirdLeaderIndex] = false
	disconnect(allRaft[(thirdLeaderIndex+1)%len(allRaft)])
	isConnect[(thirdLeaderIndex+1)%len(allRaft)] = false
	time.Sleep(500 * time.Millisecond)
	forthLeaderIndex, _ := CheckOneLeader(allRaft, isConnect)
	if forthLeaderIndex != -1 {
		t.Fatalf("Want Only One leader")
	}

	reConnect(allRaft[thirdLeaderIndex])
	isConnect[thirdLeaderIndex] = true
	time.Sleep(700 * time.Millisecond)
	fifthLeaderIndex, _ := CheckOneLeader(allRaft, isConnect)
	if fifthLeaderIndex == -1 {
		t.Fatalf("Want Only One leader")
	}

	reConnect(allRaft[(thirdLeaderIndex+1)%len(allRaft)])
	isConnect[(thirdLeaderIndex+1)%len(allRaft)] = true
	sixthLeaderIndex, _ := CheckOneLeader(allRaft, isConnect)
	if sixthLeaderIndex == -1 {
		t.Fatalf("Want Only One leader")
	}
	StopRaftNode(allRaft)

}

func TestMultiRaftDisconnectRandom(t *testing.T) {
	rpc.SetUseRPC("grpc")
	system.RegisterAndUseDefaultClock()
	system.RegisterAndUseDefaultRandom()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	log.RegisterAndUseMockLog(mockCtrl)
	raftNum := 7
	allRaft, _, isConnect := CreateRaftNode(raftNum)
	iters := 10
	for i := 0; i < iters; i++ {
		fmt.Printf("##### %v ######\n", i)
		r1 := rand.Int() % raftNum
		r2 := rand.Int() % raftNum
		r3 := rand.Int() % raftNum
		disconnect(allRaft[r1])
		disconnect(allRaft[r2])
		disconnect(allRaft[r3])
		isConnect[r1] = false
		isConnect[r2] = false
		isConnect[r3] = false
		tmpLeaderIndex, _ := CheckOneLeader(allRaft, isConnect)
		if tmpLeaderIndex == -1 {
			t.Fatalf("Want Only One leader")
		}
		reConnect(allRaft[r1])
		reConnect(allRaft[r2])
		reConnect(allRaft[r3])
		isConnect[r1] = true
		isConnect[r2] = true
		isConnect[r3] = true
	}
	tmpLeaderIndex, _ := CheckOneLeader(allRaft, isConnect)
	if tmpLeaderIndex == -1 {
		t.Fatalf("Want Only One leader")
	}
	StopRaftNode(allRaft)
}

func TestMajorityNum(t *testing.T) {
	rpc.SetUseRPC("grpc")
	system.RegisterAndUseDefaultClock()
	system.RegisterAndUseDefaultRandom()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	log.RegisterAndUseMockLog(mockCtrl)
	raftNum := []int{3, 5, 7}
	expectNum := []int{2, 3, 4}
	for key, value := range raftNum {
		allRaft, _, _ := CreateRaftNode(value)
		for _, oneRaft := range allRaft {
			if oneRaft.clusterMajority != int32(expectNum[key]) {
				t.Fatalf("raftnum[%v] majority is [%v] want [%v]", value, oneRaft.clusterMajority, expectNum[key])
			}
			oneRaft.Stop()
		}
	}

}
