package node

import (
	"testing"
)

func TestSetUseNodeProvider(t *testing.T) {
	err := SetNodeProvider("noprovider")
	if err == nil {
		t.Fatalf("SetNodeProvider(noprovider)=nil, want !=nil")
	}
	nodeProvider, err := GetNodeProvider()
	if err == nil || nodeProvider != nil {
		t.Fatalf("GetNodeProvider()=[%v][%v], want= nil,err", nodeProvider, err)
	}

}
