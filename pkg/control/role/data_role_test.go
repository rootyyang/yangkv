package role

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rootyyang/yangkv/pkg/infrastructure/log"
	"github.com/rootyyang/yangkv/pkg/infrastructure/system"
)

func TestDataRole(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	log.RegisterAndUseMockLog(mockCtrl)
	system.RegisterAndUseMockClock(mockCtrl)

	dataRole := GetDataRole()
	dataRole.Start("TestClusterName", []string{"127.0.0.1:3888"}, "1888")

}
