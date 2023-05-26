package system

import "github.com/golang/mock/gomock"

func RegisterAndUseMockClock(mockCtl *gomock.Controller) *MockClock {
	clockMock := NewMockClock(mockCtl)
	retClock = clockMock
	return clockMock
}

func RegisterAndUseDefaultClock() {
	retClock = &clockDefault{}
}
