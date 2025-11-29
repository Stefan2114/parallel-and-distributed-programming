package main

import (
	"math/rand"
)

const MaxVal = 30
const KARATSUBA_CUTOFF = 64

func newPolynomial(size int) []int64 {
	polynomial := make([]int64, size)
	for i := 0; i < size; i++ {
		polynomial[i] = rand.Int63n(MaxVal)
	}
	return polynomial
}

func pad(p []int64, length int) []int64 {
	if len(p) >= length {
		return p
	}
	newP := make([]int64, length)
	copy(newP, p)
	return newP
}

func polyAdd(p, q []int64) []int64 {
	lenP, lenQ := len(p), len(q)
	res := make([]int64, max(lenP, lenQ))
	copy(res, p)
	for i, nr := range q {
		res[i] += nr
	}
	return res
}

func polySub(p, q []int64) []int64 {
	lenP, lenQ := len(p), len(q)
	res := make([]int64, max(lenP, lenQ))
	copy(res, p)
	for i, nr := range q {
		res[i] -= nr
	}
	return res
}

func combineKaratsubaResults(RHigh, RMidTerm, RLow []int64, m, n, lenP, lenQ int) []int64 {

	finalLen := lenP + lenQ - 1
	if finalLen <= 0 {
		return []int64{}
	}

	RMidSub1 := polySub(RMidTerm, RHigh)
	RMid := polySub(RMidSub1, RLow)

	resultPadded := make([]int64, m*2)

	for i, v := range RLow {
		resultPadded[i] += v
	}
	for i, v := range RMid {
		resultPadded[i+n] += v
	}
	for i, v := range RHigh {
		resultPadded[i+2*n] += v
	}

	return resultPadded[:finalLen]
}
