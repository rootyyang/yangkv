package cluster

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/infrastructure/log"
)

func TestGetLocalAndMaster(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	log.RegisterAndUseMockLog(mockCtrl)
	clusterMetaManager := GetClusterMetaManager()
	err := clusterMetaManager.Start("TestClusterName", []string{"127.0.0.1:1234"}, "1234")
	if err != nil {
		t.Fatalf("clusterMetaManager.Start=[%v], want=nil", err)
	}
	localNodeMeta := clusterMetaManager.GetLocalNodeMeta()
	if localNodeMeta.ClusterName != "TestClusterName" {
		t.Fatalf("clusterMetaManager.GetLocalNodeMeta().ClusterName=[%v], want=TestClusterName", localNodeMeta.ClusterName)
	}
	masterNodeMeta := clusterMetaManager.GetMasterNodeMeta()
	if masterNodeMeta.ClusterName != "TestClusterName" || masterNodeMeta.Ip != "127.0.0.1" || masterNodeMeta.Port != "1234" {
		t.Fatalf("clusterMetaManager.GetMasterNodeMeta()=[%v], want={TestClusterName, 127.0.0.1, 1234", masterNodeMeta)
	}

}
