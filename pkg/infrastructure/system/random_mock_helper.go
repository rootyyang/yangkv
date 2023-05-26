package system

import gomock "github.com/golang/mock/gomock"

func RegisterAndUseMockRandom(mockCtl *gomock.Controller) *MockRandom {
	randomMock := NewMockRandom(mockCtl)
	gRandom = randomMock
	return randomMock
}
func RegisterAndUseDefaultRandom() {
	gRandom = new(defaultRandom)
}
