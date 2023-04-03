package cluster

import (
	"context"
	"fmt"
	"sync"

	gomock "github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/domain/retcode"
	"github.com/rootyyang/yangkv/pkg/infrastructure/log"
	"github.com/rootyyang/yangkv/pkg/infrastructure/rpc"
	"github.com/rootyyang/yangkv/pkg/proto"
)

//默认处理函数
var gRPCServerHandle rpc.RPCSeverHandleInterface = &defaultRPCServerHandle{masterStrategy: &MasterHandleStrategy{}, dataStrategy: &DataHandleStrategy{}}

func GetRPCServerHandle() rpc.RPCSeverHandleInterface {
	return gRPCServerHandle
}

type defaultRPCServerHandle struct {
	strategy MasterOrDataHandleStrategy

	strategyMutex sync.Mutex

	masterStrategy MasterOrDataHandleStrategy
	dataStrategy   MasterOrDataHandleStrategy
}

func (df *defaultRPCServerHandle) BecomeMaster() {
	df.strategyMutex.Lock()
	defer df.strategyMutex.Unlock()
	df.strategy = df.masterStrategy
}
func (df *defaultRPCServerHandle) NoLongerMaster() {
	df.strategyMutex.Lock()
	defer df.strategyMutex.Unlock()
	df.strategy = df.dataStrategy

}
func (df *defaultRPCServerHandle) Start() error {
	_, err := GetNodeManager()
	if err != nil {
		return err
	}
	df.strategy = df.dataStrategy
	GetMasterManager().RegisterPerceivedMasterChange(df)
	return nil
}

func (df *defaultRPCServerHandle) Stop() error {
	GetMasterManager().UnRegisterPerceivedMasterChange(df)
	return nil
}

func (df *defaultRPCServerHandle) HeartBeat(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error) {
	return df.strategy.HeartBeat(ctx, req)
}

func (df *defaultRPCServerHandle) LeaveCluster(ctx context.Context, req *proto.LeaveClusterReq) (*proto.LeaveClusterResp, error) {
	return df.strategy.LeaveCluster(ctx, req)
}

type MasterOrDataHandleStrategy interface {
	HeartBeat(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error)
	LeaveCluster(ctx context.Context, req *proto.LeaveClusterReq) (*proto.LeaveClusterResp, error)
}
type MasterHandleStrategy struct {
}

func (master *MasterHandleStrategy) HeartBeat(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error) {
	nodeManager, _ := GetNodeManager()
	ret := retcode.Success
	retmsg := "Success"
	err := nodeManager.JoinCluster(req.NodeMeta)
	if err != nil {
		log.GetLogInstance().Errorf("nodeManager.JoinCluster(ip:%v, port:%v) error[%v]", req.NodeMeta.Ip, req.NodeMeta.Port, err)
		ret = retcode.NodeJoinMasterError
		retmsg = err.Error()
	}
	heartBeatResp := proto.HeartBeatResp{RetCode: ret, RetMessage: retmsg}
	return &heartBeatResp, nil
}

func (master *MasterHandleStrategy) LeaveCluster(ctx context.Context, req *proto.LeaveClusterReq) (*proto.LeaveClusterResp, error) {
	nodeManager, _ := GetNodeManager()
	nodeManager.LeftCluster(req.NodeMeta)
	LeaveClusterResp := &proto.LeaveClusterResp{RetCode: 0, RetMessage: "success"}
	return LeaveClusterResp, nil
}

type DataHandleStrategy struct {
}

func (data *DataHandleStrategy) HeartBeat(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error) {
	fmt.Println("DataHandleStrategy")
	heartBeatResp := proto.HeartBeatResp{RetCode: 0, RetMessage: "success"}
	return &heartBeatResp, nil
}

func (data *DataHandleStrategy) LeaveCluster(ctx context.Context, req *proto.LeaveClusterReq) (*proto.LeaveClusterResp, error) {
	LeaveClusterResp := &proto.LeaveClusterResp{RetCode: retcode.NotMasterError, RetMessage: "Node Not Master"}
	return LeaveClusterResp, nil
}

func RegsiterAndUseMockHandleStrategy(mockCtl *gomock.Controller) (*MockMasterOrDataHandleStrategy, *MockMasterOrDataHandleStrategy) {
	masterHandleStrategy := NewMockMasterOrDataHandleStrategy(mockCtl)
	dataHandleStrategy := NewMockMasterOrDataHandleStrategy(mockCtl)
	gRPCServerHandle = &defaultRPCServerHandle{masterStrategy: masterHandleStrategy, dataStrategy: dataHandleStrategy}
	return masterHandleStrategy, dataHandleStrategy
}
