package main

import "sync"

// A value of 64 is a common choice.
const KARATSUBA_CUTOFF = 64

func PolyMulSequential(p, q []int) []int {
	lenP := len(p)
	lenQ := len(q)
	if lenP == 0 || lenQ == 0 {
		return []int{}
	}
	resultLen := lenP + lenQ - 1
	result := make([]int, resultLen)
	for i := 0; i < lenP; i++ {
		for j := 0; j < lenQ; j++ {
			result[i+j] += p[i] * q[j]
		}
	}
	return result
}

func PolyMulKaratsuba(p, q []int) []int {
	lenP := len(p)
	lenQ := len(q)

	finalLen := lenP + lenQ - 1
	if finalLen <= 0 {
		return []int{}
	}

	if lenP < KARATSUBA_CUTOFF || lenQ < KARATSUBA_CUTOFF {
		return PolyMulSequential(p, q)
	}

	m := max(lenP, lenQ)
	if m%2 != 0 {
		m++ // Make m even
	}

	pPadded := pad(p, m)
	qPadded := pad(q, m)
	n := m / 2

	// P(X) = P1(X)*X^n + P2(X)
	p1 := pPadded[n:] // High part (P1)
	p2 := pPadded[:n] // Low part (P2)

	// Q(X) = Q1(X)*X^n + Q2(X)
	q1 := qPadded[n:] // High part (Q1)
	q2 := qPadded[:n] // Low part (Q2)

	// 1. RHigh = P1 * Q1
	RHigh := PolyMulKaratsuba(p1, q1)

	// 2. RLow = P2 * Q2
	RLow := PolyMulKaratsuba(p2, q2)

	// 3. RMidTerm = (P1 + P2) * (Q1 + Q2)
	p1p2 := polyAdd(p1, p2)
	q1q2 := polyAdd(q1, q2)
	RMidTerm := PolyMulKaratsuba(p1p2, q1q2)

	// --- The Karatsuba Trick ---
	// RMid = RMidTerm - RHigh - RLow
	RMidSub1 := polySub(RMidTerm, RHigh)
	RMid := polySub(RMidSub1, RLow)

	// --- Combine the results ---
	// Result = RHigh * X^(2n) + RMid * X^n + RLow

	// The padded length is m*2-1.
	resultPadded := make([]int, m*2)

	// Add RLow (starts at index 0)
	copy(resultPadded, RLow)

	// Add RMid (starts at index n)
	for i, v := range RMid {
		resultPadded[i+n] += v
	}

	// Add RHigh (starts at index 2*n)
	for i, v := range RHigh {
		resultPadded[i+2*n] += v
	}

	// Trim the padded result down to the correct final length.
	return resultPadded[:finalLen]
}

func polyMulKaratsubaParallel(p, q []int, nrThreads int) []int {
	sem := make(chan struct{}, nrThreads)
	return polyMulKaratsubaParallelCoarse(p, q, sem)
}

func polyMulKaratsubaParallelCoarse(p, q []int, sem chan struct{}) []int {
	lenP := len(p)
	lenQ := len(q)

	if lenP < KARATSUBA_CUTOFF || lenQ < KARATSUBA_CUTOFF {
		return PolyMulSequential(p, q)
	}

	m := max(lenP, lenQ)
	if m%2 != 0 {
		m++
	}
	pPadded := pad(p, m)
	qPadded := pad(q, m)

	n := m / 2
	p1, p2 := pPadded[n:], pPadded[:n] // High, Low
	q1, q2 := qPadded[n:], qPadded[:n] // High, Low

	// --- Parallel Execution (3 Goroutines) ---
	var wg sync.WaitGroup
	wg.Add(2)

	var RHigh, RLow, RMidTerm []int

	// RHigh = P1 * Q1
	select {
	case sem <- struct{}{}:
		// SUCCESS: We got a slot. Run in a new goroutine.
		go func() {
			RHigh = polyMulKaratsubaParallelCoarse(p1, q1, sem) // Recurse
			<-sem                                               // Release the slot
			wg.Done()
		}()
	default:
		// FAILED: Pool is full. Run sequentially in *this* goroutine.
		RHigh = polyMulKaratsubaParallelCoarse(p1, q1, sem)
		wg.Done()
	}

	// Try to acquire another slot
	// RLow = P2 * Q2

	select {
	case sem <- struct{}{}:
		// SUCCESS: We got a slot. Run in a new goroutine.
		go func() {
			RLow = polyMulKaratsubaParallelCoarse(p2, q2, sem) // Recurse
			<-sem                                              // Release the slot
			wg.Done()
		}()
	default:
		// FAILED: Pool is full. Run sequentially in *this* goroutine.
		RLow = polyMulKaratsubaParallelCoarse(p2, q2, sem)
		wg.Done()
	}

	// RMidTerm = (P1 + P2) * (Q1 + Q2)
	p1p2 := polyAdd(p1, p2)
	q1q2 := polyAdd(q1, q2)
	RMidTerm = polyMulKaratsubaParallelCoarse(p1p2, q1q2, sem)

	wg.Wait()
	return combineKaratsubaResults(RHigh, RMidTerm, RLow, m, n, lenP, lenQ)
}

// ---
// ## 3. Fine-Grained Parallel (Inefficient "3^k")
// ---
// This version spawns 2 new goroutines at *every* recursive step,
// creating an exponential number of goroutines.

func PolyMulKaratsubaParallelFine(p, q []int) []int {
	lenP := len(p)
	lenQ := len(q)

	// Base Case: Switch to simpler algorithm for small inputs
	if lenP < KARATSUBA_CUTOFF || lenQ < KARATSUBA_CUTOFF {
		return PolyMulSequential(p, q)
	}

	m := max(lenP, lenQ)
	if m%2 != 0 {
		m++
	}
	pPadded := pad(p, m)
	qPadded := pad(q, m)

	n := m / 2
	p1, p2 := pPadded[n:], pPadded[:n] // High, Low
	q1, q2 := qPadded[n:], qPadded[:n] // High, Low

	var wg sync.WaitGroup
	wg.Add(2) // We'll spawn 2 new goroutines, and do 1 on the current thread

	var RHigh, RLow, RMidTerm []int

	// Goroutine 1: RHigh = P1 * Q1
	go func() {
		defer wg.Done()
		// CRITICAL: Calls *itself* recursively
		RHigh = PolyMulKaratsubaParallelFine(p1, q1)
	}()

	// Goroutine 2: RLow = P2 * Q2
	go func() {
		defer wg.Done()
		RLow = PolyMulKaratsubaParallelFine(p2, q2)
	}()

	// Current Thread: RMidTerm = (P1 + P2) * (Q1 + Q2)
	p1p2 := polyAdd(p1, p2)
	q1q2 := polyAdd(q1, q2)
	RMidTerm = PolyMulKaratsubaParallelFine(p1p2, q1q2)

	wg.Wait()
	return combineKaratsubaResults(RHigh, RMidTerm, RLow, m, n, lenP, lenQ)
}

func combineKaratsubaResults(RHigh, RMidTerm, RLow []int, m, n, lenP, lenQ int) []int {
	// Perform the Karatsuba trick to get the middle term
	RMidSub1 := polySub(RMidTerm, RHigh)
	RMid := polySub(RMidSub1, RLow)

	// Allocate the padded result array
	resultPadded := make([]int, m*2)

	// Combine: Result = R_low + (R_mid * X^n) + (R_high * X^2n)
	copy(resultPadded, RLow)
	for i, v := range RMid {
		resultPadded[i+n] += v
	}
	for i, v := range RHigh {
		resultPadded[i+2*n] += v
	}

	// Trim the padded result to the correct final length
	finalLen := lenP + lenQ - 1
	if finalLen <= 0 {
		return []int{}
	}
	return resultPadded[:finalLen]
}

func polyAdd(p, q []int) []int {
	maxLen := max(len(p), len(q))
	res := make([]int, maxLen)
	copy(res, p)

	for i := 0; i < len(q); i++ {
		res[i] += q[i]
	}
	return res
}

func polySub(p, q []int) []int {
	maxLen := max(len(p), len(q))
	res := make([]int, maxLen)
	copy(res, p)

	for i := 0; i < len(q); i++ {
		res[i] -= q[i]
	}
	return res
}

func pad(p []int, length int) []int {
	if len(p) >= length {
		return p
	}
	newP := make([]int, length)
	copy(newP, p)
	return newP
}
