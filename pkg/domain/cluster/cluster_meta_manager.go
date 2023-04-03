package cluster

import (
	"encoding/base64"
	"net"

	"github.com/rootyyang/yangkv/pkg/infrastructure/log"
	"github.com/rootyyang/yangkv/pkg/infrastructure/system"
	"github.com/rootyyang/yangkv/pkg/proto"
)

//TODO 当前不感受集群信息的变化，之后要感受集群信息的变化
type ClusterMetaManager interface {
	Start(pClusterName string, pDiscoveryList []string, plocalNodeTransport string) error
	GetLocalNodeMeta() *proto.NodeMeta
	GetMasterNodeMeta() *proto.NodeMeta
	Stop()
}

var gClusterMetaManager ClusterMetaManager = new(clusterMetaManagerImp)

func GetClusterMetaManager() ClusterMetaManager {
	return gClusterMetaManager
}

type clusterMetaManagerImp struct {
	localNodeMeta  *proto.NodeMeta
	masterNodeMeta *proto.NodeMeta
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
	cmm.localNodeMeta = &proto.NodeMeta{NodeId: localNodeId, Ip: localNodeAddr, Port: plocalNodeTransport, ClusterName: pClusterName, ClusterId: pClusterName, IsMaster: false}

	masterNodeId := base64.StdEncoding.EncodeToString([]byte(pDiscoveryList[0]))
	cmm.masterNodeMeta = &proto.NodeMeta{NodeId: masterNodeId, Ip: masterIp, Port: masterPort, ClusterName: pClusterName, ClusterId: pClusterName, IsMaster: true}
	return nil
}
func (cmm *clusterMetaManagerImp) GetLocalNodeMeta() *proto.NodeMeta {
	return cmm.localNodeMeta
}
func (cmm *clusterMetaManagerImp) GetMasterNodeMeta() *proto.NodeMeta {
	return cmm.masterNodeMeta
}
func (cmm *clusterMetaManagerImp) Stop() {
	return
}
