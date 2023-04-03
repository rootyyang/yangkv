package system

import (
	"time"

	"github.com/golang/mock/gomock"
)

type Clock interface {
	Now() time.Time
	Ticker(d time.Duration) *time.Ticker
}

var retClock Clock = &ClockDefault{}

func GetClock() Clock {
	return retClock
}

type ClockDefault struct {
}

func (c *ClockDefault) Now() time.Time {
	return time.Now()
}

func (c *ClockDefault) Ticker(d time.Duration) *time.Ticker {
	return time.NewTicker(d)
}

//用于mock测试
func RegisterAndUseMockClock(mockCtl *gomock.Controller) *MockClock {
	clockMock := NewMockClock(mockCtl)
	retClock = clockMock
	return clockMock
}
