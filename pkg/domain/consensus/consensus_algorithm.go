package consensus

type PerceivedMasterChange interface {
	BecomeMaster()
	NoLongerMaster()
	MasterChange(pNewMaster string)
}

type ConsensusAlgorithm interface {
	RegisterPerceivedMasterChange(PerceivedMasterChange)
	UnRegisterPerceivedMasterChange(PerceivedMasterChange)
}
