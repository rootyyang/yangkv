package system

import "math/rand"

type Random interface {
	Int() int
}

type defaultRandom struct {
}

var gRandom Random = new(defaultRandom)

func GetRandom() Random {
	return gRandom
}

func (*defaultRandom) Int() int {
	return rand.Int()
}
