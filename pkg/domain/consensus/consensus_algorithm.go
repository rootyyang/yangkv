package consensus

type Command interface {
	ToString() string
	ParseFrom(string) error
}
type StateMachine interface {
	ApplyLog(pLog Command) error
}

/*
type PerceivedMasterChange interface {
	BecomeMaster()
	NoLongerMaster()
}

type ConsensusAlgorithm interface {
	RegisterPerceivedMasterChange(PerceivedMasterChange)
	UnRegisterPerceivedMasterChange(PerceivedMasterChange)
	GetAllNode() map[string]node.Node
	GetNodeByID(pID string) node.Node
}*/
