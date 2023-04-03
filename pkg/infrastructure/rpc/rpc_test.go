package rpc

import (
	"testing"

	"github.com/golang/mock/gomock"
)

func TestSetUseRPC(t *testing.T) {
	err := SetUseRPC("norpc")
	if err == nil {
		t.Fatalf("SetUseRPC(norpc)=nil, want !=nil")
	}
	rpcClientProvider, err := GetRPCClientProvider()
	if err == nil || rpcClientProvider != nil {
		t.Fatalf("GetRPCClientProvider()=[%v][%v], want= nil,err", rpcClientProvider, err)
	}
	rpcServer, err := GetRPCServer()
	if err == nil || rpcServer != nil {
		t.Fatalf("GetRPCServer()=[%v][%v], want= nil,err", rpcServer, err)
	}

}

func TestMockRPC(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockRPCCProvider, mockRPCServer := RegisterAndUseMockRPC(mockCtrl)

	getRPCClientProvider, err := GetRPCClientProvider()
	if err != nil || getRPCClientProvider == nil {
		t.Fatalf("GetRPCClientProvider()=[%v][%v], want= not nil,nil", getRPCClientProvider, err)
	}
	getRPCServer, err := GetRPCServer()
	if err != nil || getRPCServer == nil {
		t.Fatalf("GetRPCServer()=[%v][%v], want= not nil,nil", getRPCServer, err)
	}
	if getRPCClientProvider != mockRPCCProvider {
		t.Fatalf("[%v] != [%v], want= equal ", getRPCClientProvider, mockRPCCProvider)
	}
	if getRPCServer != mockRPCServer {
		t.Fatalf("[%v] != [%v], want= equal ", getRPCServer, mockRPCServer)
	}
}
