package consensus

import (
	"fmt"
	"testing"
)

type TestCommand struct {
	serial string
}

func (tc *TestCommand) ToString() string {
	return tc.serial
}
func (tc *TestCommand) ParseFrom(pSerialString string) error {
	tc.serial = pSerialString
	return nil
}

func TestRaftLogBasicFunction(t *testing.T) {
	bufLengths := []uint32{3, 10, 29, 100}
	for _, len := range bufLengths {
		rl := &raftLogs{}
		rl.init(len)
		_, _, isNull := rl.back()
		if !isNull {
			t.Fatalf("lc.back()=any, any, %v, want any, any, true", isNull)
		}
		for i := 1; i < int(len); i++ {
			_, err := rl.append(&TestCommand{fmt.Sprintf("%d", i)}, uint64(i))
			if err != nil {
				t.Fatalf("lc.insert(TestCommand(%v)) = %v , want nil", i, err)
			}
			lastIndex, lastTerm, isNull := rl.back()
			if isNull || lastIndex != uint64(i) || lastTerm != uint64(i) {
				t.Fatalf("lc.back()=%v, %v, %v, want %v, %v, true", lastIndex, lastTerm, isNull, i, i)
			}
		}
		_, err := rl.append(&TestCommand{fmt.Sprintf("%d", len)}, uint64(len))
		if err == nil {
			t.Fatalf("lc.insert(TestCommand(%v)) = nil , want not nil", len)
		}
		//term<back, index<back
		upToDateLast := rl.upToDateLast(uint64(len-2), uint64(len-2))
		if upToDateLast {
			t.Fatalf("lc.upToDateLast(%v, %v) = true , want false", len-2, len-2)
		}
		//term<back, index=back
		upToDateLast = rl.upToDateLast(uint64(len-2), uint64(len-1))
		if upToDateLast {
			t.Fatalf("lc.upToDateLast(%v, %v) = true , want false", len-2, len-1)
		}
		//term<back, index>back
		upToDateLast = rl.upToDateLast(uint64(len-2), uint64(len))
		if upToDateLast {
			t.Fatalf("lc.upToDateLast(%v, %v) = true , want false", len-2, len)
		}
		//term=back, index<back
		upToDateLast = rl.upToDateLast(uint64(len-1), uint64(len-2))
		if upToDateLast {
			t.Fatalf("lc.upToDateLast(%v, %v) = true , want false", len-1, len-2)
		}
		//term=back, index=back
		upToDateLast = rl.upToDateLast(uint64(len-1), uint64(len-1))
		if !upToDateLast {
			t.Fatalf("lc.upToDateLast(%v, %v) = true , want false", len-1, len-1)
		}

		//term=back, index>back
		upToDateLast = rl.upToDateLast(uint64(len-1), uint64(len))
		if !upToDateLast {
			t.Fatalf("lc.upToDateLast(%v, %v) = true , want false", len-1, len)
		}
		//term>back, index<back
		upToDateLast = rl.upToDateLast(uint64(len), uint64(len-2))
		if !upToDateLast {
			t.Fatalf("lc.upToDateLast(%v, %v) = true , want false", len, len-2)
		}
		//term>back, index=back
		upToDateLast = rl.upToDateLast(uint64(len), uint64(len-1))
		if !upToDateLast {
			t.Fatalf("lc.upToDateLast(%v, %v) = true , want false", len, len-1)
		}
		//term>back, index>back
		upToDateLast = rl.upToDateLast(uint64(len), uint64(len))
		if !upToDateLast {
			t.Fatalf("lc.upToDateLast(%v, %v) = true , want false", len, len)
		}

	}
}
