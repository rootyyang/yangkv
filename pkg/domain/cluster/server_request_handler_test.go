package cluster

import (
	"context"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/domain/retcode"
	"github.com/rootyyang/yangkv/pkg/proto"
)

func TestServerRequestHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	RegisterAndUseMockNodeManager(mockCtrl)

	mockMasterStrategy, mockDataStrategy := RegsiterAndUseMockHandleStrategy(mockCtrl)

	serverHandle := GetRPCServerHandle()

	err := serverHandle.Start()
	if err != nil {
		t.Fatalf("DefaultRPCServerHandle.Start()=[%v] wan=nil", err)
	}

	GetMasterManager().BecomeMaster()
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	heartBeatReq := proto.HeartBeatReq{NodeMeta: &proto.NodeMeta{NodeId: "testNodeId", ClusterName: "testClusterName", Ip: "127.0.0.1", Port: "6888"}}
	heartBeatResp := proto.HeartBeatResp{RetCode: 0, RetMessage: "success"}
	mockMasterStrategy.EXPECT().HeartBeat(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error) {
		if req.NodeMeta.String() != heartBeatReq.NodeMeta.String() {
			t.Errorf("mockRPCClient.HeartBeat(any, [%v]), want rpcServerHandle.HeartBeat(any, [%v]])", req, heartBeatReq)
		}
		return nil, nil
	}).Return(&heartBeatResp, nil).Times(1)

	serverHandle.HeartBeat(ctx, &heartBeatReq)

	leaveClusterReq := proto.LeaveClusterReq{NodeMeta: &proto.NodeMeta{NodeId: "testNodeId", ClusterName: "testClusterNameLeave", Ip: "127.0.0.1", Port: "6888"}}
	leaveClusterResp := proto.LeaveClusterResp{RetCode: 0, RetMessage: "success"}
	mockMasterStrategy.EXPECT().LeaveCluster(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *proto.LeaveClusterReq) (*proto.LeaveClusterResp, error) {
		if req.NodeMeta.String() != leaveClusterReq.NodeMeta.String() {
			t.Errorf("mockRPCClient.LeaveCluster(any, [%v]), want mockRPCClient.LeaveCluster(any, [%v]])", req, leaveClusterReq)
		}
		return nil, nil
	}).Return(&leaveClusterResp, nil).Times(1)
	serverHandle.LeaveCluster(ctx, &leaveClusterReq)

	GetMasterManager().BecomeMaster()
	mockMasterStrategy.EXPECT().HeartBeat(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error) {
		if req.NodeMeta.String() != heartBeatReq.NodeMeta.String() {
			t.Errorf("mockRPCClient.HeartBeat(any, [%v]), want rpcServerHandle.HeartBeat(any, [%v]])", req, heartBeatReq)
		}
		return nil, nil
	}).Return(&heartBeatResp, nil).Times(1)
	serverHandle.HeartBeat(ctx, &heartBeatReq)
	mockMasterStrategy.EXPECT().LeaveCluster(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *proto.LeaveClusterReq) (*proto.LeaveClusterResp, error) {
		if req.NodeMeta.String() != leaveClusterReq.NodeMeta.String() {
			t.Errorf("mockRPCClient.LeaveCluster(any, [%v]), want mockRPCClient.LeaveCluster(any, [%v]])", req, leaveClusterReq)
		}
		return nil, nil
	}).Return(&leaveClusterResp, nil).Times(1)

	serverHandle.LeaveCluster(ctx, &leaveClusterReq)

	GetMasterManager().NoLogerMaster()

	mockDataStrategy.EXPECT().HeartBeat(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error) {
		if req.NodeMeta.String() != heartBeatReq.NodeMeta.String() {
			t.Errorf("mockRPCClient.HeartBeat(any, [%v]), want rpcServerHandle.HeartBeat(any, [%v]])", req, heartBeatReq)
		}
		return nil, nil
	}).Return(&heartBeatResp, nil).Times(1)
	serverHandle.HeartBeat(ctx, &heartBeatReq)

	mockDataStrategy.EXPECT().LeaveCluster(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *proto.LeaveClusterReq) (*proto.LeaveClusterResp, error) {
		if req.NodeMeta.String() != leaveClusterReq.NodeMeta.String() {
			t.Errorf("mockRPCClient.LeaveCluster(any, [%v]), want mockRPCClient.LeaveCluster(any, [%v]])", req, leaveClusterReq)
		}
		return nil, nil
	}).Return(&leaveClusterResp, nil).Times(1)
	serverHandle.LeaveCluster(ctx, &leaveClusterReq)

	GetMasterManager().BecomeMaster()
	mockMasterStrategy.EXPECT().HeartBeat(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error) {
		if req.NodeMeta.String() != heartBeatReq.NodeMeta.String() {
			t.Errorf("mockRPCClient.HeartBeat(any, [%v]), want rpcServerHandle.HeartBeat(any, [%v]])", req, heartBeatReq)
		}
		return nil, nil
	}).Return(&heartBeatResp, nil).Times(1)
	serverHandle.HeartBeat(ctx, &heartBeatReq)
	mockMasterStrategy.EXPECT().LeaveCluster(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *proto.LeaveClusterReq) (*proto.LeaveClusterResp, error) {
		if req.NodeMeta.String() != leaveClusterReq.NodeMeta.String() {
			t.Errorf("mockRPCClient.LeaveCluster(any, [%v]), want mockRPCClient.LeaveCluster(any, [%v]])", req, leaveClusterReq)
		}
		return nil, nil
	}).Return(&leaveClusterResp, nil).Times(1)

	serverHandle.LeaveCluster(ctx, &leaveClusterReq)

}

func TestMasterHandleStrategy(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockNodeManager := RegisterAndUseMockNodeManager(mockCtrl)

	masterHandle := MasterHandleStrategy{}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	heartBeatReq := proto.HeartBeatReq{NodeMeta: &proto.NodeMeta{NodeId: "testNodeId", ClusterName: "testClusterName", Ip: "127.0.0.1", Port: "6888"}}

	leaveClusterReq := proto.LeaveClusterReq{NodeMeta: &proto.NodeMeta{NodeId: "testNodeId", ClusterName: "testClusterNameLeave", Ip: "127.0.0.1", Port: "6888"}}

	mockNodeManager.EXPECT().JoinCluster(heartBeatReq.NodeMeta).Return(nil).Times(1)
	getHeartBeatResp, err := masterHandle.HeartBeat(ctx, &heartBeatReq)
	if getHeartBeatResp.RetCode != 0 || err != nil {
		t.Errorf("masterHandle.HeartBeat()=[%v][%v], want={0, success}, nil", getHeartBeatResp, err)
	}

	mockNodeManager.EXPECT().LeftCluster(leaveClusterReq.NodeMeta).Return(nil).Times(1)
	getLeaveResp, err := masterHandle.LeaveCluster(ctx, &leaveClusterReq)
	if getLeaveResp.RetCode != 0 || err != nil {
		t.Errorf("masterHandle.LeaveCluster()=[%v][%v], want={0, success}, nil", getLeaveResp, err)
	}

}

func TestDataHandleStrategy(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dataHandle := DataHandleStrategy{}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	heartBeatReq := proto.HeartBeatReq{NodeMeta: &proto.NodeMeta{NodeId: "testNodeId", ClusterName: "testClusterName", Ip: "127.0.0.1", Port: "6888"}}

	leaveClusterReq := proto.LeaveClusterReq{NodeMeta: &proto.NodeMeta{NodeId: "testNodeId", ClusterName: "testClusterNameLeave", Ip: "127.0.0.1", Port: "6888"}}
	getHeartBeatResp, err := dataHandle.HeartBeat(ctx, &heartBeatReq)
	if getHeartBeatResp.RetCode != 0 || err != nil {
		t.Errorf("masterHandle.HeartBeat()=[%v][%v], want={0, success}, nil", getHeartBeatResp, err)
	}

	getLeaveResp, err := dataHandle.LeaveCluster(ctx, &leaveClusterReq)
	if getLeaveResp.RetCode != retcode.NotMasterError || err != nil {
		t.Errorf("masterHandle.HeartBeat()=[%v][%v], want={0, success}, nil", getLeaveResp, err)
	}

}
