package consensus

import (
	"container/heap"
	"sort"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	test := [][]int{{10, 8, 9, 7, 6, 10}, {4, 2, 5, 6, 7}}
	for _, oneSlice := range test {
		var pq priorityQueue
		heap.Init(&pq)
		for _, oneValue := range oneSlice {
			heap.Push(&pq, &insertLogResp{0, uint64(oneValue), 0, nil, nil})
		}
		sort.Ints(oneSlice)
		for _, oneValue := range oneSlice {
			if uint64(oneValue) != heap.Pop(&pq).(*insertLogResp).prevIndex {
				t.Fatalf("oneSlice[%v] != priorityQueue[%v]", oneValue, pq.buffers)
			}
		}
	}
}
