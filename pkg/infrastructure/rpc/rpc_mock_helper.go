package rpc

import gomock "github.com/golang/mock/gomock"

func RegisterAndUseMockRPC(mockCtl *gomock.Controller) (*MockRPCClientProvider, *MockRPCServerProvider) {
	MockRPCServerProvider := NewMockRPCServerProvider(mockCtl)
	mockRPCProvider := NewMockRPCClientProvider(mockCtl)
	registerRPC("mockRPC", mockRPCProvider, MockRPCServerProvider)
	SetUseRPC("mockRPC")
	return mockRPCProvider, MockRPCServerProvider
}
