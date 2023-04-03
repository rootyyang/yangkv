package log

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/golang/mock/gomock"
)

const (
	DebugLevel = iota
	WarnLevel
	ErrorLevel
)

type LogInterface interface {
	Start(pFileName string, pMaxSizeMB int, pMaxBackupFileNum int, pMaxAgeDay int, pCompress bool) error
	Sync() error
	Stop() error
	SetLevel(pLevel int) error
	Errorf(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Debugf(template string, args ...interface{})
}

type logProvider struct {
	name2Log map[string]LogInterface
	global   LogInterface
}

var globalLog = logProvider{name2Log: make(map[string]LogInterface)}

//该函数不可并发访问
func SetUseLog(pLogName string) error {
	if log, ok := globalLog.name2Log[pLogName]; ok {
		globalLog.global = log
		return nil
	}
	return fmt.Errorf("Log[%s] Not Found", pLogName)
}

//该函数可以并发访问
func GetLog() (LogInterface, error) {
	if globalLog.global == nil {
		return nil, fmt.Errorf("Please call SetUseLog(pLogName string) first")
	}
	return globalLog.global, nil
}

//方便使用，不足od是否SetUseLog的判断
func GetLogInstance() LogInterface {
	return globalLog.global
}

//该函数不可并发访问
func registerLog(pLogName string, pLogInterface LogInterface) {
	globalLog.name2Log[pLogName] = pLogInterface
}

func RegisterAndUseMockLog(mockCtl *gomock.Controller) {
	logInterface := NewMockLogInterface(mockCtl)
	logInterface.EXPECT().Debugf(gomock.Any(), gomock.Any()).Do(func(template string, args ...interface{}) {
		_, file, line, _ := runtime.Caller(6)
		fmt.Printf(file+":"+strconv.Itoa(line)+":"+template+"\n", args...)
	}).AnyTimes()
	logInterface.EXPECT().Warnf(gomock.Any(), gomock.Any()).Do(func(template string, args ...interface{}) {
		_, file, line, _ := runtime.Caller(6)
		fmt.Printf(file+":"+strconv.Itoa(line)+":"+template+"\n", args...)
	}).AnyTimes()
	logInterface.EXPECT().Errorf(gomock.Any(), gomock.Any()).Do(func(template string, args ...interface{}) {
		_, file, line, _ := runtime.Caller(6)
		fmt.Printf(file+":"+strconv.Itoa(line)+":"+template+"\n", args...)
	}).AnyTimes()
	registerLog("mockLog", logInterface)
	SetUseLog("mockLog")
}
