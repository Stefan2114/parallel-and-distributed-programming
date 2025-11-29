package main

import (
	"math/big"
	"math/rand"
)

const MaxVal = 30000000
const KARATSUBA_CUTOFF = 64

func newPolynomial(size int) []*big.Int {
	polynomial := make([]*big.Int, size)
	for i := 0; i < size; i++ {
		polynomial[i] = big.NewInt(rand.Int63n(MaxVal))
	}
	return polynomial
}

func pad(p []*big.Int, length int) []*big.Int {
	if len(p) >= length {
		return p
	}
	newP := make([]*big.Int, length)
	copy(newP, p)
	for i := len(p); i < length; i++ {
		newP[i] = big.NewInt(0)
	}
	return newP
}

func polyAdd(p, q []*big.Int) []*big.Int {
	lenP, lenQ := len(p), len(q)
	maxLen := max(lenP, lenQ)
	res := make([]*big.Int, maxLen)
	for i := 0; i < maxLen; i++ {
		res[i] = big.NewInt(0)
		if i < lenP {
			res[i].Add(res[i], p[i])
		}
		if i < lenQ {
			res[i].Add(res[i], q[i])
		}
	}
	return res
}

func polySub(p, q []*big.Int) []*big.Int {
	lenP, lenQ := len(p), len(q)
	maxLen := max(lenP, lenQ)
	res := make([]*big.Int, maxLen)
	for i := 0; i < maxLen; i++ {
		res[i] = big.NewInt(0)
		if i < lenP {
			res[i].Add(res[i], p[i])
		}
		if i < lenQ {
			res[i].Sub(res[i], q[i])
		}
	}
	return res
}

func combineKaratsubaResults(RHigh, RMidTerm, RLow []*big.Int, m, n, lenP, lenQ int) []*big.Int {
	RMidSub1 := polySub(RMidTerm, RHigh)
	RMid := polySub(RMidSub1, RLow)

	resultPadded := make([]*big.Int, m*2)
	for i := range resultPadded {
		resultPadded[i] = big.NewInt(0)
	}

	for i, v := range RLow {
		resultPadded[i].Add(resultPadded[i], v)
	}
	for i, v := range RMid {
		resultPadded[i+n].Add(resultPadded[i+n], v)
	}
	for i, v := range RHigh {
		resultPadded[i+2*n].Add(resultPadded[i+2*n], v)
	}

	finalLen := lenP + lenQ - 1
	if finalLen <= 0 {
		return []*big.Int{}
	}
	return resultPadded[:finalLen]
}
