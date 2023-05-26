package role

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/domain/cluster"
	"github.com/rootyyang/yangkv/pkg/domain/node"
	"github.com/rootyyang/yangkv/pkg/domain/retcode"
	"github.com/rootyyang/yangkv/pkg/infrastructure/log"
	"github.com/rootyyang/yangkv/pkg/infrastructure/rpc"
	"github.com/rootyyang/yangkv/pkg/proto"
)

func TestMasterRole(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	log.RegisterAndUseMockLog(mockCtrl)
	err := node.SetNodeProvider("rpcnode")
	if err != nil {
		t.Fatalf("node.SetNodeProvider(rpcnode)=%v, want nil", err)
	}
	err = cluster.SetNodeManager("default")
	if err != nil {
		t.Fatalf("node.SetNodeManager(default)=%v, want nil", err)
	}
	err = rpc.SetUseRPC("grpc")
	if err != nil {
		t.Fatalf("node.SetUseRPC(grpc)=%v, want nil", err)
	}

	nodeManager, err := cluster.GetNodeManager()
	if err != nil || nodeManager == nil {
		t.Fatalf("GetNodeManager()=[%v][%v], want= not nil, nil ", nodeManager, err)
	}

	//mockNodeProvider := node.RegisterAndUseMockNode(mockCtrl)
	masterRole := GetMasterRole()
	err = masterRole.Start("TestClusterName", []string{"127.0.0.1:3888"}, "3888")
	if err != nil {
		t.Fatalf("masterRole.Start().IsMaster(TestClusterName, []string{127.0.0.1:3888}, 3888)=[%v], want=nil ", err)
	}

	if cluster.GetMasterManager().IsMaster() != true {
		t.Fatalf("cluster.GetMasterManager().IsMaster()=false, want=true ")
	}
	masterNodeMeta := cluster.GetClusterMetaManager().GetMasterNodeMeta()

	nodeProvier, err := node.GetNodeProvider()
	if err != nil || nodeProvier == nil {
		t.Fatalf("node.GetNodeProvider()=%v,%v , want not nil,  nil", nodeProvier, err)
	}
	type nodeWrapper struct {
		node     node.Node
		nodeMeta *proto.NodeMeta
	}

	nodeMap := make(map[string]nodeWrapper)
	for i := 0; i < 5; i++ {
		nodeMeta := &proto.NodeMeta{NodeId: fmt.Sprintf("node%v", i), ClusterName: "testClusterName", Ip: "127.0.0.1", Port: fmt.Sprintf("123%v", i)}
		nodeClient := nodeProvier.GetNode(masterNodeMeta)
		if nodeClient == nil {
			t.Fatalf("nodeProvier.GetNode(%v)=nil , want not nil", masterNodeMeta)
		}
		err = nodeClient.Start()
		if err != nil {
			t.Fatalf("nodeClient.Start()=%v , want nil", err)
		}

		nodeMap[nodeMeta.NodeId] = nodeWrapper{nodeClient, nodeMeta}
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	nodeWantMap := make(map[string]interface{})
	for key, value := range nodeMap {
		nodeWantMap[key] = nil
		resp, err := value.node.HeartBeat(ctx, &proto.HeartBeatReq{NodeMeta: value.nodeMeta})
		if err != nil || resp == nil || resp.RetCode != retcode.Success {
			t.Fatalf("node.HeartBeat()=[%v][%v], want=success, nil ", resp, err)
		}
		nodeGetMap := nodeManager.GetAllNode()
		if len(nodeGetMap) != len(nodeWantMap) {
			t.Fatalf("nodeManager.GetAllNode()=[%v], want=[%v] ", nodeGetMap, nodeWantMap)
		}
		for key, _ := range nodeGetMap {
			if _, ok := nodeWantMap[key]; !ok {
				t.Fatalf("nodeManager.GetAllNode()=[%v], want=[%v] ", nodeGetMap, nodeWantMap)
			}
		}
	}

	for key, value := range nodeMap {
		delete(nodeWantMap, key)
		resp, err := value.node.LeaveCluster(ctx, &proto.LeaveClusterReq{NodeMeta: value.nodeMeta})
		if err != nil || resp == nil || resp.RetCode != retcode.Success {
			t.Fatalf("node.HeartBeat()=[%v][%v], want=success, nil ", resp, err)
		}
		nodeGetMap := nodeManager.GetAllNode()
		if len(nodeGetMap) != len(nodeWantMap) {
			t.Fatalf("nodeManager.GetAllNode()=[%v], want=[%v] ", nodeGetMap, nodeWantMap)
		}
		for key, _ := range nodeGetMap {
			if _, ok := nodeWantMap[key]; !ok {
				t.Fatalf("nodeManager.GetAllNode()=[%v], want=[%v] ", nodeGetMap, nodeWantMap)
			}
		}
	}

}
