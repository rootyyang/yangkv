package config

import "testing"

func TestSetConfig(t *testing.T) {
	err := SetConfig("noconfig")
	if err == nil {
		t.Fatalf("SetConfig(noconfig)=nil, want !=nil")
	}
	ConfigInterface, err := GetConfig()
	if err == nil || ConfigInterface != nil {
		t.Fatalf("GetConfig()=[%v][%v], want= nil,err", ConfigInterface, err)
	}
}
