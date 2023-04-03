package cluster

import (
	"context"
	"fmt"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/domain/node"
	"github.com/rootyyang/yangkv/pkg/infrastructure/log"
	"github.com/rootyyang/yangkv/pkg/infrastructure/system"
	"github.com/rootyyang/yangkv/pkg/proto"
)

func TestNodeWithHeartBeat(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	remoteNode := node.NewMockNodeInterface(mockCtrl)

	mockClock := system.RegisterAndUseMockClock(mockCtrl)
	nwhb := nodeWithHeartBeat{}
	cancelContext, _ := context.WithCancel(context.Background())

	localNodeMetat := &proto.NodeMeta{NodeId: "localNode", ClusterName: "testClusterName", Ip: "127.0.0.1", Port: "1234"}

	remoteTickerChannel := make(chan time.Time)
	remoteNode.EXPECT().Start().Return(nil).Times(1)
	mockClock.EXPECT().Ticker(gomock.Any()).Return(&time.Ticker{C: remoteTickerChannel}).Times(1)
	err := nwhb.Start(localNodeMetat, remoteNode, cancelContext)
	if err != nil {
		t.Fatalf("nodeWithHeartBeat.Start()=[%v], want=nil", err)
	}
	var i int32 = 1
	for ; i < 5; i++ {
		remoteNode.EXPECT().HeartBeat(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("Test Error")).Times(1)

		remoteTickerChannel <- time.Now()
		time.Sleep(time.Millisecond * 50)
		if nwhb.GetHeartBeatFailuresNum() != i {
			t.Fatalf("GetHeartBeatFailuresNum()=[%v], want=[%v]", nwhb.GetHeartBeatFailuresNum(), i)
		}
	}

	remoteNode.EXPECT().HeartBeat(gomock.Any(), gomock.Any()).Return(&proto.HeartBeatResp{RetCode: 0, RetMessage: "success"}, nil).Times(1)
	remoteTickerChannel <- time.Now()
	time.Sleep(time.Millisecond * 50)
	if nwhb.GetHeartBeatFailuresNum() != 0 {
		t.Fatalf("GetHeartBeatFailuresNum()=[%v], want=0", nwhb.GetHeartBeatFailuresNum())
	}
}

func TestNodeHeartBeat(t *testing.T) {
	err := SetNodeManager("default")
	if err != nil {
		t.Fatalf("SetNodeManager(default)=[%v], want=nil", err)
	}
	nodeManager, err := GetNodeManager()
	if err != nil || nodeManager == nil {
		t.Fatalf("GetNodeManager()=[%v][%v], want= not nil, nil ", nodeManager, err)
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	log.RegisterAndUseMockLog(mockCtrl)
	mockNodeProvider := node.RegisterAndUseMockNode(mockCtrl)
	mockClock := system.RegisterAndUseMockClock(mockCtrl)
	nodeManagerChannel := make(chan time.Time)

	mockClock.EXPECT().Ticker(gomock.Any()).Return(&time.Ticker{C: nodeManagerChannel}).Times(1)
	localNodeMetat := &proto.NodeMeta{NodeId: "localNode", ClusterName: "testClusterName", Ip: "127.0.0.1", Port: "1234"}
	nodeManager.Init(localNodeMetat)
	err = nodeManager.Start()
	if err != nil {
		t.Fatalf("nodeManager.Start()=[%v], want= not nil ", err)
	}
	defer nodeManager.Stop()

	type mockWrapper struct {
		mockNode *node.MockNodeInterface
		meta     *proto.NodeMeta
		channel  chan time.Time
	}
	nodeMap := make(map[string]mockWrapper)
	for i := 0; i < 5; i++ {
		nodeMeta := &proto.NodeMeta{NodeId: fmt.Sprintf("node%v", i), ClusterName: "testClusterName", Ip: "127.0.0.1", Port: fmt.Sprintf("123%v", i)}
		node := node.NewMockNodeInterface(mockCtrl)
		nodeMap[nodeMeta.NodeId] = mockWrapper{node, nodeMeta, make(chan time.Time)}

	}

	nodeWantMap := make(map[string]node.NodeInterface)

	for key, value := range nodeMap {
		mockNodeProvider.EXPECT().GetNode(nodeMap[key].meta).Return(value.mockNode).Times(1)
		value.mockNode.EXPECT().GetNodeMeta().Return(nodeMap[key].meta).AnyTimes()
		value.mockNode.EXPECT().Start().Return(nil).Times(1)
		mockClock.EXPECT().Ticker(gomock.Any()).Return(&time.Ticker{C: value.channel}).Times(1)
		nodeManager.JoinCluster(nodeMap[key].meta)
		time.Sleep(time.Millisecond * 10)

		nodeWantMap[key] = value.mockNode

		nodeGetMap := nodeManager.GetAllNode()
		if len(nodeGetMap) != len(nodeWantMap) {
			t.Fatalf("nodeManager.GetAllNode()=[%v], want=[%v] ", nodeGetMap, nodeWantMap)
		}
		for key, getValue := range nodeGetMap {
			if wantValue, ok := nodeWantMap[key]; !ok && getValue != wantValue {
				t.Fatalf("nodeManager.GetAllNode()=[%v], want=[%v] ", nodeGetMap, nodeWantMap)
			}
		}
	}
	for _, value := range nodeMap {
		value.mockNode.EXPECT().HeartBeat(gomock.Any(), gomock.Any()).Return(&proto.HeartBeatResp{RetCode: -1, RetMessage: "error"}, nil)
		value.channel <- time.Now()
		value.mockNode.EXPECT().HeartBeat(gomock.Any(), gomock.Any()).Return(&proto.HeartBeatResp{RetCode: 0, RetMessage: "error"}, fmt.Errorf("error"))
		value.channel <- time.Now()
		nodeManagerChannel <- time.Now()
		time.Sleep(time.Millisecond * 10)
	}

	//重新比对一次
	nodeGetMap := nodeManager.GetAllNode()
	if len(nodeGetMap) != len(nodeWantMap) {
		t.Fatalf("nodeManager.GetAllNode()=[%v], want=[%v] ", nodeGetMap, nodeWantMap)
	}
	for key, getValue := range nodeGetMap {
		if wantValue, ok := nodeWantMap[key]; !ok && getValue != wantValue {
			t.Fatalf("nodeManager.GetAllNode()=[%v], want=[%v] ", nodeGetMap, nodeWantMap)
		}
	}

	for key, value := range nodeMap {
		value.mockNode.EXPECT().HeartBeat(gomock.Any(), gomock.Any()).Return(&proto.HeartBeatResp{RetCode: -1, RetMessage: "error"}, nil)
		value.mockNode.EXPECT().Stop().Times(1)
		value.channel <- time.Now()
		time.Sleep(time.Millisecond * 10)
		nodeManagerChannel <- time.Now()
		time.Sleep(time.Millisecond * 10)
		delete(nodeWantMap, key)
		nodeGetMap := nodeManager.GetAllNode()
		if len(nodeGetMap) != len(nodeWantMap) {
			t.Fatalf("nodeManager.GetAllNode()=[%v], want=[%v]", nodeGetMap, nodeWantMap)
		}
		for key, getValue := range nodeGetMap {
			if wantValue, ok := nodeWantMap[key]; !ok && getValue != wantValue {
				t.Fatalf("nodeManager.GetAllNode()=[%v], want=[%v] ", nodeGetMap, nodeWantMap)
			}
		}
	}

}
