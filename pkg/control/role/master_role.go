package role

import (
	"fmt"

	"github.com/rootyyang/yangkv/pkg/domain/cluster"
	"github.com/rootyyang/yangkv/pkg/infrastructure/log"
	"github.com/rootyyang/yangkv/pkg/infrastructure/rpc"
)

//交给一个对象，来组合领域对象
type MasterRole interface {
	Start(pClusterName string, pDiscoveryList []string, plocalNodeTransport string) error
	Stop()
}

var gMasterRole MasterRole = new(masterRoleImp)

func GetMasterRole() MasterRole {
	return gMasterRole
}

type masterRoleImp struct {
}

func (mr *masterRoleImp) Start(pClusterName string, pDiscoveryList []string, plocalNodeTransport string) error {
	nodeManager, err := cluster.GetNodeManager()
	if err != nil || nodeManager == nil {
		fmt.Printf("cluster.GetNodeManager() error %s", err.Error())
		return err
	}
	gRequestHandle := cluster.GetRPCServerHandle()
	gRequestHandle.Start()
	cluster.GetMasterManager().BecomeMaster()
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
	cluster.GetClusterMetaManager().Start(pClusterName, pDiscoveryList, plocalNodeTransport)
	nodeManager.Init(cluster.GetClusterMetaManager().GetLocalNodeMeta())

	err = nodeManager.Start()
	if err != nil {
		log.GetLogInstance().Errorf("nodeManager.Start() Error %v", err)
		return err
	}
	return nil
}
func (mr *masterRoleImp) Stop() {
	nodeManager, _ := cluster.GetNodeManager()
	nodeManager.Stop()
	gRequestHandle := cluster.GetRPCServerHandle()
	gRequestHandle.Stop()
}
