package node

import (
	"fmt"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/infrastructure/system"
	"github.com/rootyyang/yangkv/pkg/proto"
)

func TestNodeWithHeartBeatImp(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	remoteNode := NewMockNodeInterface(mockCtrl)
	heartBeatDetector := NewMockPerceivedNodeHeartBeatFail(mockCtrl)
	mockClock := system.RegisterAndUseMockClock(mockCtrl)
	nwhb := GetNodeWithHeartBeatProvider().GetNode()
	localNodeMetat := &proto.NodeMeta{NodeId: "localNode", ClusterName: "testClusterName", Ip: "127.0.0.1", Port: "1234"}

	remoteTickerChannel := make(chan time.Time)
	mockClock.EXPECT().Ticker(gomock.Any()).Return(&time.Ticker{C: remoteTickerChannel}).Times(1)

	err := nwhb.Start(localNodeMetat, remoteNode)
	if err != nil {
		t.Fatalf("nodeWithHeartBeat.Start()=[%v], want=nil", err)
	}
	nwhb.RegisterHeartBeatFailMoreThanThree(heartBeatDetector)

	for i := 1; i < 5; i++ {
		if i > 2 {
			heartBeatDetector.EXPECT().HeartBeatFail(remoteNode, i).Times(1)
		}
		remoteNode.EXPECT().HeartBeat(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("Test Error")).Times(1)
		remoteTickerChannel <- time.Now()
		time.Sleep(time.Millisecond * 10)
	}
}
