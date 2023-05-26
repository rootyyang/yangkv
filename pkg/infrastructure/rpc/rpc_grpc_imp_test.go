package rpc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetClientAndServer(t *testing.T, pIp, pPort string, pServerHandler RaftServiceHandler) (RPClient, RPCServer) {
	err := SetUseRPC("grpc")
	if err != nil {
		t.Fatalf("SetUseRPC(norpc) != nil, want=nil")
	}
	rpcClientProvider, err := GetRPCClientProvider()
	if err != nil || rpcClientProvider == nil {
		t.Fatalf("GetRPCClientProvider()=[%v][%v], want= not nil ,nil", rpcClientProvider, err)
	}
	rpcServerProvider, err := GetRPCServerProvider()
	if err != nil || rpcServerProvider == nil {
		t.Fatalf("GetRPCServer()=[%v][%v], want= not nil, nil", rpcServerProvider, err)
	}
	rpcClient := rpcClientProvider.GetRPCClient(pIp, pPort)
	rpcServer := rpcServerProvider.GetRPCServer(pPort)
	rpcServer.RegisterRaftHandle(pServerHandler)
	err = rpcServer.Start()
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
	rpcServerHandle := NewMockRaftServiceHandler(mockCtrl)
	rpcClient, rpcServer := GetClientAndServer(t, "127.0.0.1", "1234", rpcServerHandle)

	defer rpcClient.Stop()
	defer rpcServer.Stop()

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	appendEntiresReq := AppendEntiresReq{LeaderTerm: 10, LeaderId: "LeaderID"}
	appendEntiresResp := AppendEntiresResp{Term: 8, Success: true}

	rpcServerHandle.EXPECT().AppendEntires(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *AppendEntiresReq) (*AppendEntiresResp, error) {
		if req.LeaderId != appendEntiresReq.LeaderId || req.LeaderTerm != appendEntiresReq.LeaderTerm {
			t.Errorf("rpcServerHandle.AppendEntires(any, [%v]), want rpcServerHandle.AppendEntires(any, appendEntiresReq)", req)
		}
		return nil, nil
	}).Return(&appendEntiresResp, nil).Times(1)
	rpcresp, err := rpcClient.AppendEntires(ctx, &appendEntiresReq)
	if err != nil {
		t.Fatalf("rpcClient.HeartBeat()=any,[%v], want=any, nil", err)
	}
	if rpcresp.Success != appendEntiresResp.Success || rpcresp.Term != appendEntiresResp.Term {
		t.Fatalf("rpcresp[%v] != appendEntiresResp[%v], want equal", *rpcresp, appendEntiresResp)
	}

}

func TestGRPCTimeout(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	rpcServerHandle := NewMockRaftServiceHandler(mockCtrl)

	rpcClient, rpcServer := GetClientAndServer(t, "127.0.0.1", "1234", rpcServerHandle)
	defer rpcServer.Stop()

	ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
	appendEntiresReq := AppendEntiresReq{LeaderTerm: 10, LeaderId: "LeaderID"}
	appendEntiresResp := AppendEntiresResp{Term: 8, Success: true}

	waitChannel := make(chan struct{})
	rpcServerHandle.EXPECT().AppendEntires(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *AppendEntiresReq) (*AppendEntiresResp, error) {
		if req.LeaderId != appendEntiresReq.LeaderId || req.LeaderTerm != appendEntiresReq.LeaderTerm {
			t.Errorf("rpcServerHandle.AppendEntires(any, [%v]), want rpcServerHandle.AppendEntires(any, appendEntiresReq)", req)
		}
		timer := time.NewTimer(1 * time.Minute)
		select {
		case <-timer.C:
			t.Fatal("ctx.Done() != close want close")
		case <-ctx.Done():
		}
		waitChannel <- struct{}{}
		return &appendEntiresResp, nil
	}).Times(1)
	rpcresp, err := rpcClient.AppendEntires(ctx, &appendEntiresReq)
	if err == nil || status.Code(err) != codes.DeadlineExceeded || rpcresp != nil {
		t.Fatalf("rpcClient.HeartBeat()=[%v],[%v], want=any, codes.DeadlineExceeded", rpcresp, err)
	}
	<-waitChannel
}

func TestGRPCCancel(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	rpcServerHandle := NewMockRaftServiceHandler(mockCtrl)

	rpcClient, rpcServer := GetClientAndServer(t, "127.0.0.1", "1234", rpcServerHandle)

	defer rpcClient.Stop()
	defer rpcServer.Stop()

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)

	appendEntiresReq := AppendEntiresReq{LeaderTerm: 10, LeaderId: "LeaderID"}
	appendEntiresResp := AppendEntiresResp{Term: 8, Success: true}
	rpcServerHandle.EXPECT().AppendEntires(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *AppendEntiresReq) (*AppendEntiresResp, error) {
		if req.LeaderId != appendEntiresReq.LeaderId || req.LeaderTerm != appendEntiresReq.LeaderTerm {
			t.Errorf("rpcServerHandle.AppendEntires(any, [%v]), want rpcServerHandle.AppendEntires(any, appendEntiresReq)", req)
		}
		tickr := time.NewTicker(2 * time.Second)
		select {
		case <-tickr.C:
			t.Fatal("ctx.Done() != close want close")
		case <-ctx.Done():
		}
		return &appendEntiresResp, nil
	}).Times(1)

	go func() {
		time.Sleep(time.Second)
		cancelFunc()
	}()
	rpcresp, err := rpcClient.AppendEntires(ctx, &appendEntiresReq)
	if err == nil || status.Code(err) != codes.Canceled {
		t.Errorf("rpcClient.HeartBeat()=[%v],[%v], want=any, codes.Canceled", rpcresp, err)
	}
}

func TestGRPCReturnError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	rpcServerHandle := NewMockRaftServiceHandler(mockCtrl)
	rpcClient, rpcServer := GetClientAndServer(t, "127.0.0.1", "1234", rpcServerHandle)

	defer rpcClient.Stop()
	defer rpcServer.Stop()

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	appendEntiresReq := AppendEntiresReq{LeaderTerm: 10, LeaderId: "LeaderID"}
	rpcServerHandle.EXPECT().AppendEntires(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *AppendEntiresReq) (*AppendEntiresResp, error) {
		if req.LeaderId != appendEntiresReq.LeaderId || req.LeaderTerm != appendEntiresReq.LeaderTerm {
			t.Errorf("rpcServerHandle.AppendEntires(any, [%v]), want rpcServerHandle.AppendEntires(any, appendEntiresReq)", req)
		}
		return nil, fmt.Errorf("Test Error")
	}).Times(1)

	rpcresp, err := rpcClient.AppendEntires(ctx, &appendEntiresReq)
	if err == nil || rpcresp != nil {
		t.Errorf("rpcClient.HeartBeat()=[%v][%v], want=nil,[%v]", rpcresp, err.Error(), fmt.Errorf("Test Error"))
	}
}

func TestGRPCRPCAfterServerStart(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	rpcServerHandle := NewMockRaftServiceHandler(mockCtrl)

	err := SetUseRPC("grpc")
	if err != nil {
		t.Fatalf("SetUseRPC(norpc) != nil, want=nil")
	}
	rpcClientProvider, err := GetRPCClientProvider()
	if err != nil || rpcClientProvider == nil {
		t.Fatalf("GetRPCClientProvider()=[%v][%v], want= not nil ,nil", rpcClientProvider, err)
	}
	rpcServerProvider, err := GetRPCServerProvider()
	if err != nil || rpcServerProvider == nil {
		t.Fatalf("GetRPCServer()=[%v][%v], want= not nil, nil", rpcServerProvider, err)
	}
	rpcClient := rpcClientProvider.GetRPCClient("127.0.0.1", "1234")
	rpcServer := rpcServerProvider.GetRPCServer("1234")
	rpcServer.RegisterRaftHandle(rpcServerHandle)

	err = rpcClient.Start()
	if err != nil {
		t.Fatalf("rpcClient.Start()=[%v], want=nil", err)
	}

	defer rpcClient.Stop()

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	appendEntiresReq := AppendEntiresReq{LeaderTerm: 10, LeaderId: "LeaderID"}
	appendEntiresResp := AppendEntiresResp{Term: 8, Success: true}
	rpcServerHandle.EXPECT().AppendEntires(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *AppendEntiresReq) (*AppendEntiresResp, error) {
		if req.LeaderId != appendEntiresReq.LeaderId || req.LeaderTerm != appendEntiresReq.LeaderTerm {
			t.Errorf("rpcServerHandle.AppendEntires(any, [%v]), want rpcServerHandle.AppendEntires(any, appendEntiresReq)", req)
		}
		return &appendEntiresResp, nil
	}).Times(1)

	rpcresp, err := rpcClient.AppendEntires(ctx, &appendEntiresReq)
	if err == nil || rpcresp != nil {
		t.Errorf("rpcClient.AppendEntires()=[%v][%v], want=nil,[%v]", rpcresp, err.Error(), fmt.Errorf("Test Error"))
	}
	err = rpcServer.Start()
	if err != nil {
		t.Fatalf("rpcServer.Start(1234, mock)=[%v], want=nil", err)
	}
	defer rpcServer.Stop()
	time.Sleep(20 * time.Millisecond)
	ctx, _ = context.WithTimeout(context.Background(), time.Second*10)
	rpcresp, err = rpcClient.AppendEntires(ctx, &appendEntiresReq)
	if err != nil || rpcresp == nil {
		t.Errorf("rpcClient.AppendEntires()=[%v][%v], want=not nil, nil ", rpcresp, err.Error())
	}

}

func TestGRPCStop(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	rpcServerHandle := NewMockRaftServiceHandler(mockCtrl)
	rpcClient, rpcServer := GetClientAndServer(t, "127.0.0.1", "1234", rpcServerHandle)

	defer rpcClient.Stop()
	defer rpcServer.Stop()

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	appendEntiresReq := AppendEntiresReq{LeaderTerm: 10, LeaderId: "LeaderID"}
	appendEntiresResp := AppendEntiresResp{Term: 8, Success: true}

	rpcServerHandle.EXPECT().AppendEntires(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *AppendEntiresReq) (*AppendEntiresResp, error) {
		if req.LeaderId != appendEntiresReq.LeaderId || req.LeaderTerm != appendEntiresReq.LeaderTerm {
			t.Errorf("rpcServerHandle.AppendEntires(any, [%v]), want rpcServerHandle.AppendEntires(any, appendEntiresReq)", req)
		}
		return nil, nil
	}).Return(&appendEntiresResp, nil).Times(1)
	rpcresp, err := rpcClient.AppendEntires(ctx, &appendEntiresReq)
	if err != nil {
		t.Fatalf("rpcClient.AppendEntires()=any,[%v], want=any, nil", err)
	}
	if rpcresp.Success != appendEntiresResp.Success || rpcresp.Term != appendEntiresResp.Term {
		t.Fatalf("rpcresp[%v] != appendEntiresResp[%v], want equal", *rpcresp, appendEntiresResp)
	}
	rpcClient.Stop()
	rpcresp, err = rpcClient.AppendEntires(ctx, &appendEntiresReq)
	if err == nil || rpcresp != nil {
		t.Fatalf("rpcClient.AppendEntires()=any,[%v], want=nil,not nil", err)
	}
	rpcClient.Start()
	rpcServerHandle.EXPECT().AppendEntires(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *AppendEntiresReq) (*AppendEntiresResp, error) {
		if req.LeaderId != appendEntiresReq.LeaderId || req.LeaderTerm != appendEntiresReq.LeaderTerm {
			t.Errorf("rpcServerHandle.AppendEntires(any, [%v]), want rpcServerHandle.AppendEntires(any, appendEntiresReq)", req)
		}
		return nil, nil
	}).Return(&appendEntiresResp, nil).Times(1)
	rpcresp, err = rpcClient.AppendEntires(ctx, &appendEntiresReq)
	if err != nil {
		t.Fatalf("rpcClient.AppendEntires()=any,[%v], want=any, nil", err)
	}
	if rpcresp.Success != appendEntiresResp.Success || rpcresp.Term != appendEntiresResp.Term {
		t.Fatalf("rpcresp[%v] != appendEntiresResp[%v], want equal", *rpcresp, appendEntiresResp)
	}
	rpcServer.Stop()
	rpcresp, err = rpcClient.AppendEntires(ctx, &appendEntiresReq)
	if err == nil || rpcresp != nil {
		t.Fatalf("rpcClient.AppendEntires()=any,[%v], want=nil,not nil", err)
	}
	rpcServer.Start()
	rpcServerHandle.EXPECT().AppendEntires(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, req *AppendEntiresReq) (*AppendEntiresResp, error) {
		if req.LeaderId != appendEntiresReq.LeaderId || req.LeaderTerm != appendEntiresReq.LeaderTerm {
			t.Errorf("rpcServerHandle.AppendEntires(any, [%v]), want rpcServerHandle.AppendEntires(any, appendEntiresReq)", req)
		}
		return nil, nil
	}).Return(&appendEntiresResp, nil).Times(1)
	time.Sleep(time.Second)
	rpcresp, err = rpcClient.AppendEntires(ctx, &appendEntiresReq)
	if err != nil {
		t.Fatalf("rpcClient.AppendEntires()=any,[%v], want=any, nil", err)
	}
	if rpcresp.Success != appendEntiresResp.Success || rpcresp.Term != appendEntiresResp.Term {
		t.Fatalf("rpcresp[%v] != appendEntiresResp[%v], want equal", *rpcresp, appendEntiresResp)
	}
}

func TestGRPCStopTwoTimes(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	rpcServerHandle := NewMockRaftServiceHandler(mockCtrl)
	_, rpcServer := GetClientAndServer(t, "127.0.0.1", "1234", rpcServerHandle)
	rpcServer.Stop()
	rpcServer.Stop()

}
