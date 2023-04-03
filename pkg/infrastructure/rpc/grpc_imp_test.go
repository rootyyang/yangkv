package rpc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetClientAndServer(t *testing.T, pIp, pPort string, pServerHandler RPCSeverHandleInterface) (RPClient, RPCServer) {
	err := SetUseRPC("grpc")
	if err != nil {
		t.Fatalf("SetUseRPC(norpc) != nil, want=nil")
	}
	rpcClientProvider, err := GetRPCClientProvider()
	if err != nil || rpcClientProvider == nil {
		t.Fatalf("GetRPCClientProvider()=[%v][%v], want= not nil ,nil", rpcClientProvider, err)
	}
	rpcServer, err := GetRPCServer()
	if err != nil || rpcServer == nil {
		t.Fatalf("GetRPCServer()=[%v][%v], want= not nil, nil", rpcServer, err)
	}
	rpcClient := rpcClientProvider.GetRPCClient(pIp, pPort)

	err = rpcServer.Start(pPort, pServerHandler)
	if err != nil {
		t.Fatalf("rpcServer.Start(1234, mock)=[%v], want=nil", err)
	}
	err = rpcClient.Start()
	if err != nil {
		t.Fatalf("rpcClient.Start(1234, mock)=[%v], want=nil", err)
	}

	return rpcClient, rpcServer
}

func TestGRPCNormal(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	rpcServerHandle := NewMockRPCSeverHandleInterface(mockCtrl)
	rpcClient, rpcServer := GetClientAndServer(t, "127.0.0.1", "1234", rpcServerHandle)

	defer rpcClient.Stop()
	defer rpcServer.Stop()

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	heartBeatReq := proto.HeartBeatReq{NodeMeta: &proto.NodeMeta{NodeId: "testNodeId", ClusterName: "testClusterName", Ip: "127.0.0.1", Port: "1234"}}
	heartBeatResp := proto.HeartBeatResp{RetCode: 0, RetMessage: "success"}

	rpcServerHandle.EXPECT().HeartBeat(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error) {
		if req.NodeMeta.String() != heartBeatReq.NodeMeta.String() {
			t.Errorf("rpcServerHandle.HeartBeat(any, [%v]), want rpcServerHandle.HeartBeat(any, heartBeatReq)", req)
		}
		return nil, nil
	}).Return(&heartBeatResp, nil).Times(1)
	rpcresp, err := rpcClient.HeartBeat(ctx, &heartBeatReq)
	if err != nil {
		t.Fatalf("rpcClient.HeartBeat()=any,[%v], want=any, nil", err)
	}
	if rpcresp.RetCode != heartBeatResp.RetCode || rpcresp.RetMessage != heartBeatResp.RetMessage {
		t.Fatalf("rpcresp[%v] != heartbeatresp[%v], want equal", rpcresp.String(), heartBeatResp.String())
	}

}

func TestGRPCTimeout(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	rpcServerHandle := NewMockRPCSeverHandleInterface(mockCtrl)

	rpcClient, rpcServer := GetClientAndServer(t, "127.0.0.1", "1234", rpcServerHandle)
	defer rpcServer.Stop()

	ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
	heartBeatReq := proto.HeartBeatReq{NodeMeta: &proto.NodeMeta{NodeId: "testNodeId", ClusterName: "testClusterName", Ip: "127.0.0.1", Port: "1234"}}
	heartBeatResp := proto.HeartBeatResp{RetCode: 0, RetMessage: "success"}

	waitChannel := make(chan struct{})
	rpcServerHandle.EXPECT().HeartBeat(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *proto.HeartBeatReq) (hbr *proto.HeartBeatResp, err error) {
		if req.NodeMeta.String() != heartBeatReq.NodeMeta.String() {
			t.Errorf("rpcServerHandle.HeartBeat(any, [%v]), want rpcServerHandle.HeartBeat(any, heartBeatReq)", req)
		}
		tickr := time.NewTicker(1 * time.Minute)
		select {
		case <-tickr.C:
			t.Fatal("ctx.Done() != close want close")
		case <-ctx.Done():
		}
		hbr, err = &heartBeatResp, nil
		waitChannel <- struct{}{}
		return
	}).Times(1)
	rpcresp, err := rpcClient.HeartBeat(ctx, &heartBeatReq)
	if err == nil || status.Code(err) != codes.DeadlineExceeded || rpcresp != nil {
		t.Fatalf("rpcClient.HeartBeat()=[%v],[%v], want=any, codes.DeadlineExceeded", rpcresp, err)
	}
	<-waitChannel
}

func TestGRPCCancel(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	rpcServerHandle := NewMockRPCSeverHandleInterface(mockCtrl)

	rpcClient, rpcServer := GetClientAndServer(t, "127.0.0.1", "1234", rpcServerHandle)

	defer rpcClient.Stop()
	defer rpcServer.Stop()

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)

	heartBeatReq := proto.HeartBeatReq{NodeMeta: &proto.NodeMeta{NodeId: "testNodeId", ClusterName: "testClusterName", Ip: "127.0.0.1", Port: "1234"}}
	heartBeatResp := proto.HeartBeatResp{RetCode: 0, RetMessage: "success"}
	rpcServerHandle.EXPECT().HeartBeat(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *proto.HeartBeatReq) (hbr *proto.HeartBeatResp, err error) {
		if req.NodeMeta.String() != heartBeatReq.NodeMeta.String() {
			t.Errorf("rpcServerHandle.HeartBeat(any, [%v]), want rpcServerHandle.HeartBeat(any, heartBeatReq)", req)
		}
		tickr := time.NewTicker(2 * time.Second)
		select {
		case <-tickr.C:
			t.Fatal("ctx.Done() != close want close")
		case <-ctx.Done():
		}
		hbr, err = &heartBeatResp, nil
		return
	}).Times(1)

	go func() {
		time.Sleep(time.Second)
		cancelFunc()
	}()
	rpcresp, err := rpcClient.HeartBeat(ctx, &heartBeatReq)
	if err == nil || status.Code(err) != codes.Canceled {
		t.Errorf("rpcClient.HeartBeat()=[%v],[%v], want=any, codes.Canceled", rpcresp, err)
	}
}

func TestGRPCReturnError(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	rpcServerHandle := NewMockRPCSeverHandleInterface(mockCtrl)
	rpcClient, rpcServer := GetClientAndServer(t, "127.0.0.1", "1234", rpcServerHandle)

	defer rpcClient.Stop()
	defer rpcServer.Stop()

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	heartBeatReq := proto.HeartBeatReq{NodeMeta: &proto.NodeMeta{NodeId: "testNodeId", ClusterName: "testClusterName", Ip: "127.0.0.1", Port: "1234"}}
	rpcServerHandle.EXPECT().HeartBeat(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *proto.HeartBeatReq) (hbr *proto.HeartBeatResp, err error) {
		if req.NodeMeta.String() != heartBeatReq.NodeMeta.String() {
			t.Errorf("rpcServerHandle.HeartBeat(any, [%v]), want rpcServerHandle.HeartBeat(any, heartBeatReq)", req)
		}
		return nil, fmt.Errorf("Test Error")
	}).Times(1)

	rpcresp, err := rpcClient.HeartBeat(ctx, &heartBeatReq)
	if err == nil || rpcresp != nil {
		t.Errorf("rpcClient.HeartBeat()=[%v][%v], want=nil,[%v]", rpcresp, err.Error(), fmt.Errorf("Test Error"))
	}
}
