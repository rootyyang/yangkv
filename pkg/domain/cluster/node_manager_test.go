package cluster

import "testing"

func TestSetNotRegistedNodeManager(t *testing.T) {
	err := SetNodeManager("noNodeManager")
	if err == nil {
		t.Fatalf("SetNodeManager(noNodeManager)=nil, want !=nil")
	}
	nodeManager, err := GetNodeManager()
	if err == nil || nodeManager != nil {
		t.Fatalf("GetNodeManager()=[%v][%v], want= nil,err", nodeManager, err)
	}

}

func TestSetRegistedNodeManager(t *testing.T) {
	err := SetNodeManager("default")
	if err != nil {
		t.Fatalf("SetNodeManager(noNodeManager)=[%v], want=nil", err)
	}
	nodeManager, err := GetNodeManager()
	if err != nil || nodeManager == nil {
		t.Fatalf("GetNodeManager()=[%v][%v], want= not nil, nil ", nodeManager, err)
	}

}
