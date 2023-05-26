package cluster

import (
	"encoding/base64"
	"net"

	"github.com/rootyyang/yangkv/pkg/domain/node"
	"github.com/rootyyang/yangkv/pkg/infrastructure/log"
	"github.com/rootyyang/yangkv/pkg/infrastructure/system"
)

//TODO 当前不感受集群信息的变化，之后要感受集群信息的变化
type ClusterMetaManager interface {
	Start(pClusterName string, pDiscoveryList []string, plocalNodeTransport string) error
	GetLocalNodeMeta() *node.NodeMeta
	GetMasterNodeMeta() *node.NodeMeta
	Stop()
}

var gClusterMetaManager ClusterMetaManager = new(clusterMetaManagerImp)

func GetClusterMetaManager() ClusterMetaManager {
	return gClusterMetaManager
}

type clusterMetaManagerImp struct {
	localNodeMeta  *node.NodeMeta
	masterNodeMeta *node.NodeMeta
}

func (cmm *clusterMetaManagerImp) Start(pClusterName string, pDiscoveryList []string, plocalNodeTransport string) error {
	localNodeAddr, err := system.GetLocalAddr()
	if err != nil {
		log.GetLogInstance().Errorf("system.GetLocalAddr() error %s", err.Error())
		return err
	}
	masterIp, masterPort, err := net.SplitHostPort(pDiscoveryList[0])
	if err != nil {
		log.GetLogInstance().Errorf("net.SplitHostPort(%v) error %s", pDiscoveryList[0], err.Error())
		return err
	}
	localNodeId := base64.StdEncoding.EncodeToString([]byte(localNodeAddr + ":" + plocalNodeTransport))
	cmm.localNodeMeta = &node.NodeMeta{NodeId: localNodeId, Ip: localNodeAddr, Port: plocalNodeTransport, ClusterName: pClusterName, ClusterId: pClusterName}

	masterNodeId := base64.StdEncoding.EncodeToString([]byte(pDiscoveryList[0]))
	cmm.masterNodeMeta = &node.NodeMeta{NodeId: masterNodeId, Ip: masterIp, Port: masterPort, ClusterName: pClusterName, ClusterId: pClusterName}
	return nil
}
func (cmm *clusterMetaManagerImp) GetLocalNodeMeta() *node.NodeMeta {
	return cmm.localNodeMeta
}
func (cmm *clusterMetaManagerImp) GetMasterNodeMeta() *node.NodeMeta {
	return cmm.masterNodeMeta
}
func (cmm *clusterMetaManagerImp) Stop() {
	return
}
