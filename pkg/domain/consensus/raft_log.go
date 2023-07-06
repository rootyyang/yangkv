package consensus

import (
	"fmt"

	"github.com/rootyyang/yangkv/pkg/infrastructure/log"
)

/*
	该对象不做并发控制这里不做并发控制，由raft对象统一控制
	为了在Append获取PrevIndex和PrevTerm，应该保证buffer中至少有一个日志
	(endPosition+1) % len(lc.buffers) == beginPostion 则表示缓存满，因此缓存实际可以存放数据数量是len(lc.buffers) - 1
*/
type raftLogs struct {
	buffers            []oneRaftLog
	beginPosition      int
	endPosition        int //最后一个日志的后一位
	commitPosition     int
	beginPositionIndex uint64
	nextAppendIndex    uint64
	//lastApplied        uint64 先跟commitIndex合并为一，如果是异步apply，则需要lastApplied
	commitIndex  uint64
	stateMachine StateMachine
}

type oneRaftLog struct {
	term          uint64
	command       Command
	commandSerial string
}

// 实际数据，从index=1 term=1开始， 日志[index=1]的prevIndex=0, prevTerm=0
func (lc *raftLogs) init(pBufferLen uint32) error {
	if pBufferLen < 2 {
		return EParamError
	}
	lc.buffers = make([]oneRaftLog, pBufferLen)
	lc.beginPositionIndex = 0
	lc.endPosition = 1
	lc.nextAppendIndex = 1
	return nil
}

// 当内存满的时候直接返回失败
func (lc *raftLogs) append(pOneLog oneRaftLog, pCurrentTerm uint64) (rMsgIndex uint64, rPrevIndex uint64, rPrevTerm uint64, rErr error) {
	if (lc.endPosition+1)%len(lc.buffers) == lc.beginPosition {
		rErr = ESpaceNotEnough
		return
	}
	var backPosition int
	if lc.endPosition == 0 {
		backPosition = len(lc.buffers) - 1
	} else {
		backPosition = lc.endPosition - 1
	}
	rPrevIndex = lc.nextAppendIndex - 1
	rPrevTerm = lc.buffers[backPosition].term
	rMsgIndex = lc.nextAppendIndex
	lc.buffers[lc.endPosition] = pOneLog
	lc.endPosition = (lc.endPosition + 1) % len(lc.buffers)
	lc.nextAppendIndex++
	return
}
func (lc *raftLogs) index2position(pIndex uint64) (rPosition int, rErr error) {
	if pIndex < lc.beginPositionIndex {
		rErr = EIndexLessThanMin
		return
	}
	if pIndex >= lc.nextAppendIndex {
		rErr = EIndexGreaterThanMax
		return
	}
	rPosition = (int(pIndex-lc.beginPositionIndex) + lc.beginPosition) % len(lc.buffers)
	return
}
func (lc *raftLogs) remainCapacity() int {
	if lc.endPosition > lc.beginPosition {
		return len(lc.buffers) - lc.endPosition + lc.beginPosition - 1
	}
	return len(lc.buffers) - -lc.beginPosition + lc.endPosition - 1

}

func (lc *raftLogs) thereIsEnoughSpace(pPrevIndex uint64, pWantStoreNum int) error {
	if pPrevIndex < lc.beginPositionIndex {
		return EIndexLessThanMin
	}
	if pPrevIndex >= lc.nextAppendIndex {
		return EIndexGreaterThanMax
	}
	if int(lc.nextAppendIndex-pPrevIndex-1)+lc.remainCapacity() < pWantStoreNum {
		return ESpaceNotEnough
	}
	return nil
}

// TODO 考虑快照的问题
// @param
// @return nil, EIndexLessThanMin, EIndexGreaterThanMax, EPrevIndexAndPrevTermNotMatch,ESpaceNotEnough

func (lc *raftLogs) insert(pPrevTerm uint64, pPrevIndex uint64, logs []oneRaftLog, pCurrentTerm uint64) error {
	//理论上，不可能收到小于自身的日志
	prevIndexPos, err := lc.index2position(pPrevIndex)
	if err != nil {
		if err == EIndexLessThanMin {
			//不可能出现这种情况
			log.Instance().Errorf("Fatal : pPrevIndex[%v] < lc.beginPositionIndex[%v]", pPrevIndex, lc.beginPosition)
		}
		return err
	}

	if lc.buffers[prevIndexPos].term != pPrevTerm {
		return EPrevIndexAndPrevTermNotMatch
	}

	if lc.thereIsEnoughSpace(pPrevIndex, len(logs)) != nil {
		return ESpaceNotEnough
	}
	beginCopyPos := prevIndexPos + 1
	for _, value := range logs {
		if beginCopyPos >= lc.endPosition {
			//覆盖
			lc.endPosition = (lc.endPosition + 1) % len(lc.buffers)
		}
		lc.buffers[beginCopyPos] = value
		beginCopyPos = (beginCopyPos + 1) % len(lc.buffers)
	}
	return nil
}

func (lc *raftLogs) upToDateLast(pLogTerm uint64, pLogIndex uint64) bool {
	lastLogIndex, lastLogTerm, isNull := lc.back()
	if isNull {
		return true
	}
	if pLogTerm > lastLogTerm {
		return true
	} else if pLogTerm == lastLogTerm {
		if pLogIndex >= lastLogIndex {
			return true
		}
		return false
	}
	return false
}
func (lc *raftLogs) back() (rLastLogIndex uint64, rLastLogTerm uint64, rIsNull bool) {
	if lc.endPosition == lc.beginPosition {
		rIsNull = true
		return
	}
	rLastLogIndex = lc.nextAppendIndex
	rLastLogTerm = lc.buffers[lc.endPosition-1].term
	return
}

//更新rAfterEndIndex
func (lc *raftLogs) get(pBeginIndex uint64, pNum int) (rPrevTerm uint64, rLogs []string, rAfterEndIndex uint64, rErr error) {
	beginPos, err := lc.index2position(pBeginIndex)
	if err != nil {
		rErr = err
		return
	}
	// ! 理论上不会出现这种情况，看snapshot 时会如何处理
	if beginPos == lc.beginPosition {
		rErr = EIndexEqualMin
		return
	}
	rPrevTerm = lc.buffers[(beginPos+len(lc.buffers)-1)%len(lc.buffers)].term

	numInBuffers := lc.nextAppendIndex - pBeginIndex
	if numInBuffers < uint64(pNum) {
		pNum = int(numInBuffers)
	}
	rLogs = make([]string, pNum)
	rAfterEndIndex = pBeginIndex + 1
	for i := 0; beginPos != lc.endPosition && i < pNum; i++ {
		rLogs[i] = lc.buffers[beginPos].commandSerial
		rAfterEndIndex++
	}
	return
}

func (lc *raftLogs) getCommitIndex() uint64 {
	return lc.commitIndex
}

// 如果该index是本term的第一个，则下一次尝试上一个term的第一个，否则尝试本term的第一个
func (lc *raftLogs) getNextTryWhenAppendEntiresFalse(pIndex uint64) (rNextTryIndex uint64, rErr error) {
	if pIndex == lc.beginPositionIndex {
		rNextTryIndex = pIndex
		return
	}
	indexPos, err := lc.index2position(pIndex)
	if err != nil {
		rErr = err
		return
	}
	//前一个index的Term
	beforePos := (indexPos + len(lc.buffers) - 1) % len(lc.buffers)
	findFirstTerm := lc.buffers[beforePos].term
	rNextTryIndex = pIndex
	for ; beforePos > lc.beginPosition && lc.buffers[beforePos].term == findFirstTerm; rNextTryIndex-- {
		beforePos = (beforePos + len(lc.buffers) - 1) % len(lc.buffers)
	}
	return
}
func (lc *raftLogs) commit(pCommitIndex uint64) error {
	if pCommitIndex <= lc.commitIndex {
		return nil
	}
	if pCommitIndex >= lc.nextAppendIndex {
		return fmt.Errorf("param bigger than max log index")
	}
	for lc.commitIndex <= pCommitIndex {
		lc.stateMachine.ApplyLog(lc.buffers[lc.commitPosition].command)
		lc.commitPosition++
	}
	lc.commitIndex = pCommitIndex
	return nil
}
func (lc *raftLogs) truncate() error {
	lc.beginPosition = (lc.commitPosition + 1) % len(lc.buffers)
	lc.beginPositionIndex = lc.commitIndex + 1
	return nil
}
