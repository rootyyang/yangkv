package role

import (
	"fmt"

	"github.com/rootyyang/yangkv/pkg/domain/cluster"
	"github.com/rootyyang/yangkv/pkg/domain/node"
	"github.com/rootyyang/yangkv/pkg/infrastructure/log"
	"github.com/rootyyang/yangkv/pkg/infrastructure/rpc"
)

type DataRole interface {
	Start(pClusterName string, pDiscoveryList []string, plocalNodeTransport string) error
	Stop()
}

var gDataRole DataRole = new(DataRoleImp)

func GetDataRole() DataRole {
	return gDataRole
}

type DataRoleImp struct {
	masterNode node.Node
}

func (dr *DataRoleImp) Start(pClusterName string, pDiscoveryList []string, plocalNodeTransport string) error {
	gRequestHandle := cluster.GetRPCServerHandle()
	gRequestHandle.Start()
	rpcServer, err := rpc.GetRPCServer()
	if err != nil {
		log.GetLogInstance().Errorf("rpc.GetRPCServer() Error %v", err)
		fmt.Printf("rpc.GetRPCServer() Error %v", err)
		return err
	}
	err = rpcServer.Start(plocalNodeTransport, gRequestHandle)
	if err != nil {
		log.GetLogInstance().Errorf("rpcServer.Start() Error %v", err)
		return err
	}
	clusterMetaManager := cluster.GetClusterMetaManager()
	clusterMetaManager.Start(pClusterName, pDiscoveryList, plocalNodeTransport)
	masterNodeDetector := cluster.GetMasterNodeDetector()
	nodeProvider, err := node.GetNodeProvider()
	if err != nil {
		fmt.Printf("GetNodeProvider.Start() Error %v", err)
		return err
	}
	dr.masterNode = nodeProvider.GetNode(clusterMetaManager.GetMasterNodeMeta())
	err = dr.masterNode.Start()
	if err != nil {
		fmt.Printf("masterNode.Start() Error %v", err)
		return err
	}

	err = masterNodeDetector.Start(clusterMetaManager.GetLocalNodeMeta(), dr.masterNode)
	if err != nil {
		fmt.Printf("masterNodeDetector.Start() Error %v", err)
		return err
	}
	return nil

}
func (dr *DataRoleImp) Stop() {
	cluster.GetRPCServerHandle().Stop()
	cluster.GetMasterNodeDetector().Stop()
	dr.masterNode.Stop()
}
