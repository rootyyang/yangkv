package config

import gomock "github.com/golang/mock/gomock"

func RegisterAndUseMockConfig(mockCtl *gomock.Controller) *MockConfigInterface {
	configInterface := NewMockConfigInterface(mockCtl)
	registerConfig("mockConfig", configInterface)
	SetConfig("mockConfig")
	return configInterface
}
