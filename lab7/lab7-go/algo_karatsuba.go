package main

import (
	"fmt"
	"sync"
)

func polyMulKaratsubaParallelCoarse(p, q []int64, sem chan struct{}) []int64 {
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

	var wg sync.WaitGroup
	wg.Add(2)
	var RHigh, RLow, RMidTerm []int64

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

func PolyMulDistributedKaratsuba(p, q []int64, rank, worldSize int) []int64 {

	if worldSize < 4 {
		if rank == 0 {
			fmt.Println("Warning: Karatsuba MPI now requires at least 4 processes (Master + 3 Workers). Running Sequential.")
			return PolyMulSequential(p, q)
		}
		return nil
	}

	if rank == 0 {
		return polyMulDistributedKaratsubaCoordinator(p, q)

	} else if rank == 1 {
		// Worker 1: High * High
		polyMulDistributedKaratsubaWorker(10)
		return nil

	} else if rank == 2 {
		// Worker 2: Low * Low
		polyMulDistributedKaratsubaWorker(20)

		return nil

	} else if rank == 3 {
		// Worker 3: (Low+High) * (Low+High)
		polyMulDistributedKaratsubaWorker(30)

		return nil

	} else {
		return nil
	}
}

func polyMulDistributedKaratsubaCoordinator(p, q []int64) []int64 {
	lenP := len(p)
	lenQ := len(q)

	if lenP < KARATSUBA_CUTOFF || lenQ < KARATSUBA_CUTOFF {
		var empty []int64
		SendPoly(empty, 1, 10)
		SendPoly(empty, 1, 11)
		SendPoly(empty, 2, 20)
		SendPoly(empty, 2, 21)
		SendPoly(empty, 3, 30)
		SendPoly(empty, 3, 31)

		RecvPoly(1, 12)
		RecvPoly(2, 22)
		RecvPoly(3, 32)

		return PolyMulSequential(p, q)
	}

	m := max(lenP, lenQ)
	if m%2 != 0 {
		m++
	}

	pPadded := pad(p, m)
	qPadded := pad(q, m)
	n := m / 2

	p1, p2 := pPadded[n:], pPadded[:n]
	q1, q2 := pPadded[n:], pPadded[:n]
	q1, q2 = qPadded[n:], qPadded[:n]

	p1p2 := polyAdd(p1, p2)
	q1q2 := polyAdd(q1, q2)

	// --- DISTRIBUTE WORK ---
	// 1. Send High parts to Rank 1
	SendPoly(p1, 1, 10)
	SendPoly(q1, 1, 11)

	// 2. Send Low parts to Rank 2
	SendPoly(p2, 2, 20)
	SendPoly(q2, 2, 21)

	// 3. Send Sum parts to Rank 3
	SendPoly(p1p2, 3, 30)
	SendPoly(q1q2, 3, 31)

	// --- GATHER RESULTS ---
	RHigh := RecvPoly(1, 12)
	RLow := RecvPoly(2, 22)
	RMidTerm := RecvPoly(3, 32)

	return combineKaratsubaResults(RHigh, RMidTerm, RLow, m, n, lenP, lenQ)
}

func polyMulDistributedKaratsubaWorker(tag int) []int64 {
	p := RecvPoly(0, tag)
	q := RecvPoly(0, tag+1)
	nrThreads := 16
	sem := make(chan struct{}, nrThreads)
	res := polyMulKaratsubaParallelCoarse(p, q, sem)
	SendPoly(res, 0, tag+2)
	return nil
}
