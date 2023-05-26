package node

import gomock "github.com/golang/mock/gomock"

func RegisterAndUseMockNode(mockCtl *gomock.Controller) *MockNodeProvider {
	mockNodeProvider := NewMockNodeProvider(mockCtl)
	gNodeProvider = mockNodeProvider
	return mockNodeProvider
}
