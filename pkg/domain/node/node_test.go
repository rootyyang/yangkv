package node

import (
	"context"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/infrastructure/rpc"
)

func TestUseRPCNodeNormal(t *testing.T) {
	nodeProvider := GetNodeProvider()
	if nodeProvider == nil {
		t.Fatalf("GetNodeProvider()=nil, want!= nil")
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	//获得RPC的Mock对象

	mockRPCClient := rpc.NewMockRPClient(mockCtrl)
	rpcProvider, _ := rpc.RegisterAndUseMockRPC(mockCtrl)
	rpcProvider.EXPECT().GetRPCClient("127.0.0.1", "1234").Return(mockRPCClient).Times(1)

	clientMeta := NodeMeta{NodeId: "testNodeId", ClusterName: "testClusterName", Ip: "127.0.0.1", Port: "1234"}
	clientNode := nodeProvider.GetNode(clientMeta)

	mockRPCClient.EXPECT().Start().Return(nil).Times(1)
	err := clientNode.Start()
	if err != nil {
		t.Fatalf("clientNode.Start()=[%v], want= nil", err)
	}
	appendEntiresReq := rpc.AppendEntiresReq{LeaderTerm: 10, LeaderId: "LeaderID"}
	appendEntiresResp := rpc.AppendEntiresResp{Term: 8, Success: true}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	mockRPCClient.EXPECT().AppendEntires(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *rpc.AppendEntiresReq) (*rpc.AppendEntiresResp, error) {
		if req.LeaderId != appendEntiresReq.LeaderId || req.LeaderTerm != appendEntiresReq.LeaderTerm {
			t.Errorf("rpcServerHandle.AppendEntires(any, [%v]), want rpcServerHandle.AppendEntires(any, appendEntiresReq)", req)
		}
		return nil, nil
	}).Return(&appendEntiresResp, nil).Times(1)

	rpcresp, err := clientNode.GetRPC().AppendEntires(ctx, &appendEntiresReq)
	if rpcresp.Success != appendEntiresResp.Success || rpcresp.Term != appendEntiresResp.Term {
		t.Fatalf("rpcresp[%v] != appendEntiresResp[%v], want equal", *rpcresp, appendEntiresResp)
	}

	requestVoteReq := rpc.RequestVoteReq{Term: 11, CandidateId: "LeaderID"}
	requestVoteResp := rpc.RequestVoteResp{Term: 9, VoteGranted: true}

	mockRPCClient.EXPECT().RequestVote(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *rpc.RequestVoteReq) (*rpc.RequestVoteResp, error) {
		if req.CandidateId != requestVoteReq.CandidateId || req.Term != requestVoteReq.Term {
			t.Errorf("rpcServerHandle.RequestVote(any, [%v]), want rpcServerHandle.AppendEntires(any, appendEntiresReq)", req)
		}

		return nil, nil
	}).Return(&requestVoteResp, nil).Times(1)
	requesRpcResp, err := clientNode.GetRPC().RequestVote(ctx, &requestVoteReq)
	if requesRpcResp.Term != requestVoteResp.Term || requesRpcResp.VoteGranted != requestVoteResp.VoteGranted {
		t.Fatalf("requesRpcResp[%v] != requestVoteResp[%v], want equal", *requesRpcResp, requestVoteResp)
	}

}

func TestUseGRPCNodeNormal(t *testing.T) {
	err := rpc.SetUseRPC("grpc")
	if err != nil {
		t.Fatalf("SetUseRPC(norpc) != nil, want=nil")
	}

	rpcServerProvider, err := rpc.GetRPCServerProvider()
	if err != nil || rpcServerProvider == nil {
		t.Fatalf("GetRPCServer()=[%v][%v], want= not nil, nil", rpcServerProvider, err)
	}
	rpcServer := rpcServerProvider.GetRPCServer("1234")

	nodeProvider := GetNodeProvider()
	if nodeProvider == nil {
		t.Fatalf("GetNodeProvider()=nil, want!= nil")
	}
	clientMeta := NodeMeta{NodeId: "testNodeId", ClusterName: "testClusterName", Ip: "127.0.0.1", Port: "1234"}
	clientNode := nodeProvider.GetNode(clientMeta)
	err = clientNode.Start()
	if err != nil {
		t.Fatalf("node start()=[%v], want=nil", err)
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	rpcServerHandle := rpc.NewMockRaftServiceHandler(mockCtrl)
	rpcServer.RegisterRaftHandle(rpcServerHandle)
	err = rpcServer.Start()
	if err != nil {
		t.Fatalf("rpcServer.Start()=[%v], want=nil", err)
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	appendEntiresReq := rpc.AppendEntiresReq{LeaderTerm: 10, LeaderId: "LeaderID"}
	appendEntiresResp := rpc.AppendEntiresResp{Term: 8, Success: true}

	rpcServerHandle.EXPECT().AppendEntires(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *rpc.AppendEntiresReq) (*rpc.AppendEntiresResp, error) {
		if req.LeaderId != appendEntiresReq.LeaderId || req.LeaderTerm != appendEntiresReq.LeaderTerm {
			t.Errorf("rpcServerHandle.AppendEntires(any, [%v]), want rpcServerHandle.AppendEntires(any, appendEntiresReq)", req)
		}
		return nil, nil
	}).Return(&appendEntiresResp, nil).Times(1)
	rpcresp, err := clientNode.GetRPC().AppendEntires(ctx, &appendEntiresReq)
	if err != nil {
		t.Fatalf("rpcClient.AppendEntires()=any,[%v], want=any, nil", err)
	}
	if rpcresp.Success != appendEntiresResp.Success || rpcresp.Term != appendEntiresResp.Term {
		t.Fatalf("rpcresp[%v] != appendEntiresResp[%v], want equal", *rpcresp, appendEntiresResp)
	}

}
