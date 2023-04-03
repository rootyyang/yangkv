package node

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/infrastructure/rpc"
	"github.com/rootyyang/yangkv/pkg/proto"
)

func TestUseRPCNodeNormal(t *testing.T) {
	err := SetNodeProvider("rpcnode")
	if err != nil {
		t.Fatalf("SetNodeProvider(noprovider)=nil, want !=nil")
	}
	nodeProvider, err := GetNodeProvider()
	if err != nil || nodeProvider == nil {
		t.Fatalf("GetNodeProvider()=[%v][%v], want= nil,err", nodeProvider, err)
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	//获得RPC的Mock对象
	mockRPCClient := rpc.NewMockRPClientInterface(mockCtrl)
	rpcProvider, _ := rpc.RegisterAndUseMockRPC(mockCtrl)
	rpcProvider.EXPECT().GetRPCClient("127.0.0.1", "1234").Return(mockRPCClient).Times(1)

	clientMeta := proto.NodeMeta{NodeId: "testNodeId", ClusterName: "testClusterName", Ip: "127.0.0.1", Port: "1234"}
	clientNode := nodeProvider.GetNode(&clientMeta)

	mockRPCClient.EXPECT().Start().Return(nil).Times(1)
	err = clientNode.Start()
	if err != nil {
		t.Fatalf("clientNode.Start()=[%v], want= nil", err)
	}

	heartBeatReq := proto.HeartBeatReq{NodeMeta: &proto.NodeMeta{NodeId: "testNodeId", ClusterName: "testClusterName", Ip: "127.0.0.1", Port: "6888"}}
	heartBeatResp := proto.HeartBeatResp{RetCode: 0, RetMessage: "success"}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	mockRPCClient.EXPECT().HeartBeat(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error) {
		if req.NodeMeta.String() != heartBeatReq.NodeMeta.String() {
			t.Errorf("mockRPCClient.HeartBeat(any, [%v]), want rpcServerHandle.HeartBeat(any, [%v]])", req, heartBeatReq)
		}
		return nil, nil
	}).Return(&heartBeatResp, nil).Times(1)
	getheartbeatresp, err := clientNode.HeartBeat(ctx, &heartBeatReq)
	if getheartbeatresp.RetCode != heartBeatResp.RetCode || getheartbeatresp.RetMessage != heartBeatResp.RetMessage {
		t.Fatalf("getheartbeatresp[%v] != heartbeatresp[%v], want equal", getheartbeatresp.String(), heartBeatResp.String())
	}

	leaveClusterReq := proto.LeaveClusterReq{NodeMeta: &proto.NodeMeta{NodeId: "testNodeId", ClusterName: "testClusterNameLeave", Ip: "127.0.0.1", Port: "6888"}}
	leaveClusterResp := proto.LeaveClusterResp{RetCode: 0, RetMessage: "success"}
	mockRPCClient.EXPECT().LeaveCluster(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *proto.LeaveClusterReq) (*proto.LeaveClusterResp, error) {
		if req.NodeMeta.String() != leaveClusterReq.NodeMeta.String() {
			t.Errorf("mockRPCClient.LeaveCluster(any, [%v]), want mockRPCClient.LeaveCluster(any, [%v]])", req, leaveClusterReq)
		}
		return nil, nil
	}).Return(&leaveClusterResp, nil).Times(1)
	getLeaveClusterresp, err := clientNode.LeaveCluster(ctx, &leaveClusterReq)
	if getLeaveClusterresp.RetCode != leaveClusterResp.RetCode || getLeaveClusterresp.RetMessage != leaveClusterResp.RetMessage {
		t.Fatalf("getLeaveClusterresp[%v] != leaveClusterResp[%v], want equal", getLeaveClusterresp.String(), leaveClusterResp.String())
	}

}
