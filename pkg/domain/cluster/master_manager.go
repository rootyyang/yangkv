package cluster

import (
	"sync"
)

type PerceivedMasterChange interface {
	BecomeMaster()
	NoLongerMaster()
}

type MasterManager interface {
	IsMaster() bool
	BecomeMaster()
	NoLogerMaster()
	RegisterPerceivedMasterChange(PerceivedMasterChange)
	UnRegisterPerceivedMasterChange(PerceivedMasterChange)
}

var gMasterManager = &DefaultMasterManager{isMaster: false, perceivedMasterMap: make(map[PerceivedMasterChange]interface{})}

func GetMasterManager() MasterManager {
	return gMasterManager
}

type DefaultMasterManager struct {
	isMaster                bool
	isMasterMutx            sync.RWMutex
	perceivedMasterMap      map[PerceivedMasterChange]interface{}
	perceivedMasterMapMutex sync.RWMutex
}

func (mm *DefaultMasterManager) IsMaster() bool {
	mm.isMasterMutx.RLock()
	defer mm.isMasterMutx.RUnlock()
	return mm.isMaster
}
func (mm *DefaultMasterManager) BecomeMaster() {
	mm.isMasterMutx.RLock()
	if mm.IsMaster() == true {
		mm.isMasterMutx.RUnlock()
		return
	}
	mm.isMasterMutx.RUnlock()

	mm.isMasterMutx.Lock()
	if mm.isMaster == true {
		mm.isMasterMutx.Unlock()
		return
	}
	mm.isMaster = true
	mm.isMasterMutx.Unlock()
	mm.perceivedMasterMapMutex.RLock()
	defer mm.perceivedMasterMapMutex.RUnlock()
	for key, _ := range mm.perceivedMasterMap {
		key.BecomeMaster()
	}

}
func (mm *DefaultMasterManager) NoLogerMaster() {
	if mm.IsMaster() == false {
		return
	}
	mm.isMasterMutx.Lock()
	if mm.isMaster == false {
		mm.isMasterMutx.Unlock()
		return
	}
	mm.isMaster = false
	mm.isMasterMutx.Unlock()
	mm.perceivedMasterMapMutex.RLock()
	defer mm.perceivedMasterMapMutex.RUnlock()
	for key, _ := range mm.perceivedMasterMap {
		key.NoLongerMaster()
	}
}

func (mm *DefaultMasterManager) RegisterPerceivedMasterChange(pPMC PerceivedMasterChange) {
	mm.perceivedMasterMapMutex.Lock()
	defer mm.perceivedMasterMapMutex.Unlock()
	mm.perceivedMasterMap[pPMC] = nil

}
func (mm *DefaultMasterManager) UnRegisterPerceivedMasterChange(pPMC PerceivedMasterChange) {
	mm.perceivedMasterMapMutex.Lock()
	defer mm.perceivedMasterMapMutex.Unlock()
	delete(mm.perceivedMasterMap, pPMC)
}
