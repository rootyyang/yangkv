package log

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/golang/mock/gomock"
)

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
