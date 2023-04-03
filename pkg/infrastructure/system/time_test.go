package system

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestUseMock(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClock := RegisterAndUseMockClock(mockCtrl)
	getClock := GetClock()
	if getClock != getClock {
		t.Fatalf("GetClock()=[%v], want=[%v] ", getClock, mockClock)
	}
	mockTicker := time.Ticker{C: make(<-chan time.Time)}

	mockClock.EXPECT().Ticker(gomock.Any()).Return(&mockTicker).Times(1)
	getTicker := getClock.Ticker(3 * time.Second)
	if &mockTicker != getTicker {
		t.Fatalf("getClock.Ticker()=[%v], want=[%v] ", getClock, mockClock)
	}

}
