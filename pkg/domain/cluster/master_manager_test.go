package cluster

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
)

func TestMasterManager(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockPerceivedMasterChange1 := NewMockPerceivedMasterChange(mockCtrl)
	mockPerceivedMasterChange2 := NewMockPerceivedMasterChange(mockCtrl)
	mockPerceivedMasterChange3 := NewMockPerceivedMasterChange(mockCtrl)

	masterManager := GetMasterManager()
	masterManager.RegisterPerceivedMasterChange(mockPerceivedMasterChange1)

	mockPerceivedMasterChange1.EXPECT().BecomeMaster().Times(1)
	masterManager.BecomeMaster()

	//再次执行的时候，不应该再次触发观察者
	masterManager.BecomeMaster()

	mockPerceivedMasterChange1.EXPECT().NoLongerMaster().Times(1)
	masterManager.NoLogerMaster()
	masterManager.NoLogerMaster()

	masterManager.RegisterPerceivedMasterChange(mockPerceivedMasterChange2)

	mockPerceivedMasterChange1.EXPECT().BecomeMaster().Times(1)
	mockPerceivedMasterChange2.EXPECT().BecomeMaster().Times(1)

	masterManager.BecomeMaster()
	//再次执行的时候，不应该再次触发观察者
	masterManager.BecomeMaster()

	mockPerceivedMasterChange1.EXPECT().NoLongerMaster().Times(1)
	mockPerceivedMasterChange2.EXPECT().NoLongerMaster().Times(1)
	masterManager.NoLogerMaster()
	masterManager.NoLogerMaster()

	masterManager.RegisterPerceivedMasterChange(mockPerceivedMasterChange3)

	mockPerceivedMasterChange1.EXPECT().BecomeMaster().Times(1)
	mockPerceivedMasterChange2.EXPECT().BecomeMaster().Times(1)
	mockPerceivedMasterChange3.EXPECT().BecomeMaster().Times(1)
	masterManager.BecomeMaster()
	//再次执行的时候，不应该再次触发观察者
	masterManager.BecomeMaster()

	mockPerceivedMasterChange1.EXPECT().NoLongerMaster().Times(1)
	mockPerceivedMasterChange2.EXPECT().NoLongerMaster().Times(1)
	mockPerceivedMasterChange3.EXPECT().NoLongerMaster().Times(1)
	masterManager.NoLogerMaster()
	masterManager.NoLogerMaster()

	masterManager.UnRegisterPerceivedMasterChange(mockPerceivedMasterChange1)

	mockPerceivedMasterChange2.EXPECT().BecomeMaster().Times(1)
	mockPerceivedMasterChange3.EXPECT().BecomeMaster().Times(1)

	masterManager.BecomeMaster()
	//再次执行的时候，不应该再次触发观察者
	masterManager.BecomeMaster()

	mockPerceivedMasterChange2.EXPECT().NoLongerMaster().Times(1)
	mockPerceivedMasterChange3.EXPECT().NoLongerMaster().Times(1)
	masterManager.NoLogerMaster()
	masterManager.NoLogerMaster()

	masterManager.UnRegisterPerceivedMasterChange(mockPerceivedMasterChange2)

	mockPerceivedMasterChange3.EXPECT().BecomeMaster().Times(1)

	masterManager.BecomeMaster()
	//再次执行的时候，不应该再次触发观察者
	masterManager.BecomeMaster()

	mockPerceivedMasterChange3.EXPECT().NoLongerMaster().Times(1)
	masterManager.NoLogerMaster()
	masterManager.NoLogerMaster()

	masterManager.UnRegisterPerceivedMasterChange(mockPerceivedMasterChange3)
	masterManager.BecomeMaster()
	masterManager.NoLogerMaster()
}
