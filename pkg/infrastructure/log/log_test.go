package log

import "testing"

func TestSetLog(t *testing.T) {
	err := SetUseLog("nolog")
	if err == nil {
		t.Fatalf("SetUseLog(nolog)=nil, want !=nil")
	}
	logInterface, err := GetLog()
	if err == nil || logInterface != nil {
		t.Fatalf("GetLog()=[%v][%v], want= nil,err", logInterface, err)
	}
}
