package system

import (
	"time"
)

//定义这两个接口的原因，是为了方便mock测试
type ClockTimer interface {
	Stop() bool
	Reset(d time.Duration) bool
	GetChannel() <-chan time.Time
}

type ClockTicker interface {
	Stop()
	Reset(d time.Duration)
	GetChannel() <-chan time.Time
}

type Clock interface {
	Now() time.Time
	Ticker(d time.Duration) ClockTicker
	Timer(d time.Duration) ClockTimer
}

var retClock Clock = &clockDefault{}

func GetClock() Clock {
	return retClock
}

type clockDefault struct {
}

//系统对应的Timer
type SysClockTimer struct {
	*time.Timer
}

type SysClockTicker struct {
	*time.Ticker
}

func (sys *SysClockTimer) GetChannel() <-chan time.Time {
	return sys.C
}

func (sys *SysClockTicker) GetChannel() <-chan time.Time {
	return sys.C
}

func (c *clockDefault) Now() time.Time {
	return time.Now()
}

func (c *clockDefault) Ticker(d time.Duration) ClockTicker {

	return &SysClockTicker{time.NewTicker(d)}
}

func (c *clockDefault) Timer(d time.Duration) ClockTimer {
	return &SysClockTimer{time.NewTimer(d)}
}

//用于mock测试
