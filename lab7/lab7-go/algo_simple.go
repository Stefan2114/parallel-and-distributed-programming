package main

import (
	"fmt"
	"sync"
)

func PolyMulSequential(p, q []int64) []int64 {
	lenP := len(p)
	lenQ := len(q)
	resultLen := lenP + lenQ - 1
	result := make([]int64, resultLen)
	for i := 0; i < lenP; i++ {
		if p[i] == 0 {
			continue
		}
		for j := 0; j < lenQ; j++ {
			result[i+j] += p[i] * q[j]
		}
	}
	return result
}

func PolyMulParallel(polyA, polyB []int64, start, len, nrThreads int) []int64 {

	result := make([]int64, len)
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

func computeElement(k int, polyA, polyB []int64) int64 {
	lenA := len(polyA)
	lenB := len(polyB)
	var localSum int64 = 0

	for i := 0; i < lenA; i++ {
		j := k - i
		if j >= 0 && j < lenB {
			localSum += polyA[i] * polyB[j]
		}
	}
	return localSum
}

func PolyMulDistributedSimple(p1, p2 []int64, rank, worldSize int) []int64 {

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

func polyMulDistributedSimpleCoordinator(totalLen, chunkSize, worldSize int) []int64 {
	finalResult := make([]int64, totalLen)
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

func polyMulDistributedSimpleWorker(p1, p2 []int64, rank, totalLen, chunkSize, worldSize, nrThreads int) {
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
