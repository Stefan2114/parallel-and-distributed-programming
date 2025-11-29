package main

import (
	"fmt"
	"math/big"
	"sync"
)

func PolyMulSequential(p, q []*big.Int) []*big.Int {
	lenP := len(p)
	lenQ := len(q)
	resultLen := lenP + lenQ - 1
	result := make([]*big.Int, resultLen)
	for i := range result {
		result[i] = big.NewInt(0)
	}
	term := new(big.Int)
	for i := 0; i < lenP; i++ {
		for j := 0; j < lenQ; j++ {
			term.Mul(p[i], q[j])
			result[i+j].Add(result[i+j], term)
		}
	}
	return result
}

func PolyMulParallel(polyA, polyB []*big.Int, start, len, nrThreads int) []*big.Int {

	result := make([]*big.Int, len)
	var wg sync.WaitGroup
	baseWork := len / nrThreads
	remainder := len % nrThreads
	currentStartIdx := start

	for k := 0; k < nrThreads; k++ {

		workSize := baseWork
		if k < remainder {
			workSize++
		}

		endIdx := currentStartIdx + workSize

		wg.Add(1)
		go func(startIndex, endIndex, globalOffset int) {
			defer wg.Done()
			for i := startIndex; i < endIndex; i++ {
				element := computeElement(i, polyA, polyB)
				result[i-globalOffset] = element
			}
		}(currentStartIdx, endIdx, start)
		currentStartIdx = endIdx
	}

	wg.Wait()
	return result
}

func computeElement(k int, polyA, polyB []*big.Int) *big.Int {
	lenA := len(polyA)
	lenB := len(polyB)
	localSum := big.NewInt(0)
	term := new(big.Int)

	for i := 0; i < lenA; i++ {
		j := k - i
		if j >= 0 && j < lenB {
			term.Mul(polyA[i], polyB[j])
			localSum.Add(localSum, term)
		}
	}
	return localSum
}

func PolyMulDistributedSimple(p1, p2 []*big.Int, rank, worldSize int) []*big.Int {

	startNrThreads := 16
	if worldSize < 2 {
		if rank == 0 {
			fmt.Println("Warning: Not enough processes for coordinator-worker split (need >= 2). Running only on master.")
			l := len(p1) + len(p2) - 1
			return PolyMulParallel(p1, p2, 0, l, startNrThreads)
		}
		return nil
	}

	p1 = BcastPoly(p1, 0)
	p2 = BcastPoly(p2, 0)

	lenA := len(p1)
	lenB := len(p2)
	totalLen := lenA + lenB - 1

	numWorkers := worldSize - 1
	chunkSize := totalLen / numWorkers

	if rank == 0 {
		return polyMulDistributedSimpleCoordinator(totalLen, chunkSize, worldSize)

	} else {
		polyMulDistributedSimpleWorker(p1, p2, rank, totalLen, chunkSize, worldSize, startNrThreads)
		return nil
	}
}

func polyMulDistributedSimpleCoordinator(totalLen, chunkSize, worldSize int) []*big.Int {
	finalResult := make([]*big.Int, totalLen)
	for src := 1; src < worldSize; src++ {
		part := RecvPoly(src, 0)

		workerRank := src - 1
		otherStart := workerRank * chunkSize

		for i, val := range part {
			if otherStart+i < totalLen {
				finalResult[otherStart+i] = val
			}
		}
	}
	return finalResult
}

func polyMulDistributedSimpleWorker(p1, p2 []*big.Int, rank, totalLen, chunkSize, worldSize, nrThreads int) {
	workerRank := rank - 1
	myStart := workerRank * chunkSize
	myEnd := myStart + chunkSize

	if rank == worldSize-1 {
		myEnd = totalLen
	}

	myLen := myEnd - myStart
	if myLen < 0 {
		myLen = 0
	}

	localResult := PolyMulParallel(p1, p2, myStart, myLen, nrThreads)
	SendPoly(localResult, 0, 0)
}
