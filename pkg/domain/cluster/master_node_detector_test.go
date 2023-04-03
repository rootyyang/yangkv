package cluster

import (
	"context"
	"fmt"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/domain/node"
	"github.com/rootyyang/yangkv/pkg/infrastructure/system"
	"github.com/rootyyang/yangkv/pkg/proto"
)

func TestMasterNodeDetector(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	localNode := node.NewMockNodeInterface(mockCtrl)
	remoteNode := node.NewMockNodeInterface(mockCtrl)

	mockClock := system.RegisterAndUseMockClock(mockCtrl)
	nwhb := nodeWithHeartBeat{}
	cancelContext, _ := context.WithCancel(context.Background())

	localNodeMetat := &proto.NodeMeta{NodeId: "localNode", ClusterName: "testClusterName", Ip: "127.0.0.1", Port: "1234"}
	localNode.EXPECT().GetNodeMeta().Return(localNodeMetat).AnyTimes()

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
