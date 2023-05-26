package log

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"
)

//{"level":"debug","ts":1675684622.739339,"caller":"log/log_zap_lumberjack.go:58","msg":"test debug "}
type logContext struct {
	Level  string  `json:"level"`
	Ts     float32 `json:"ts"`
	Caller string  `json:"caller"`
	Msg    string  `json:"msg"`
}

func TestWriteLog(t *testing.T) {
	err := SetUseLog("zaplumberjack")
	if err != nil {
		t.Fatalf("SetUseLog(zaplumberjack)=[%v], want ==nil", err)
	}
	logInterface, err := GetLog()
	if err != nil || logInterface == nil {
		t.Fatalf("GetLog()=[%v][%v], want= not nil, nil", logInterface, err)
	}

	err = logInterface.Start("./test.log", 100, 10, 10, false)
	if err != nil {
		t.Fatalf("Start(test,  100, 10, 10, false)=[%v], want= nil", err)
	}

	err = logInterface.SetLevel(DebugLevel)
	if err != nil {
		t.Fatalf("SetLevel(DebugLevel)=[%v], want= nil", err)
	}
	logInterface.Debugf("first")
	logInterface.Warnf("first")
	logInterface.Errorf("first")
	err = logInterface.SetLevel(WarnLevel)
	if err != nil {
		t.Fatalf("SetLevel(DebugLevel)=[%v], want= nil", err)
	}
	logInterface.Debugf("second")
	logInterface.Warnf("second")
	logInterface.Errorf("second")
	err = logInterface.SetLevel(ErrorLevel)
	if err != nil {
		t.Fatalf("SetLevel(DebugLevel)=[%v], want= nil", err)
	}
	logInterface.Debugf("third")
	logInterface.Warnf("third")
	logInterface.Errorf("third")
	logInterface.Sync()

	testFile, err := os.Open("./test.log")
	if err != nil || testFile == nil {
		t.Fatalf("os.Open(test.log)=[%v][%v], want=not nil, nil", testFile, err)
	}
	logContextSlice := make([]logContext, 0)
	inputScanner := bufio.NewScanner(testFile)
	for inputScanner.Scan() {
		context := inputScanner.Text()
		contextStruct := logContext{}
		err = json.Unmarshal([]byte(context), &contextStruct)
		if err != nil {
			t.Fatalf("json.Unmarshal([%v])=[%v], want=nil", context, err)
		}
		logContextSlice = append(logContextSlice, contextStruct)
	}
	testFile.Close()
	if len(logContextSlice) != 6 {
		t.Errorf("len(logContextSlice)=[%v], want=6", len(logContextSlice))
	}
	result := []logContext{{Level: "debug", Msg: "first"}, {Level: "warn", Msg: "first"}, {Level: "error", Msg: "first"}, {Level: "warn", Msg: "second"}, {Level: "error", Msg: "second"}, {Level: "error", Msg: "third"}}
	for i, value := range result {
		if logContextSlice[i].Level != value.Level || logContextSlice[i].Msg != value.Msg {
			t.Errorf("[%v]!=[%v], want equal", logContextSlice[i], value)
		}
	}

	err = os.Remove("./test.log")
	if err != nil {
		t.Fatalf("os.Remove(test.log)=[%v], want=nil", err)
	}
}
