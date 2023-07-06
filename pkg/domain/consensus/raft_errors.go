package consensus

import "errors"

var (
	EIndexLessThanMin             = errors.New("raftLog: Index Less than minimum")
	EIndexEqualMin                = errors.New("raftLog: Index equal to minimum")
	EIndexGreaterThanMax          = errors.New("raftLog: Index greater than maximum")
	EPrevIndexAndPrevTermNotMatch = errors.New("raftLog: Previous index and previous term not match")
	ETimeOut                      = errors.New("Time Out")
	EThereIsNewLeader             = errors.New("There is a new master node")
	ESpaceNotEnough               = errors.New("There is not enough space")
	EParamError                   = errors.New("Param Error")
)
