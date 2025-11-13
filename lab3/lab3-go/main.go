package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const MaxVal = 10

type Matrix [][]int

func computeElement(A, B, C Matrix, row, col, size, threadID int) {
	//fmt.Printf("Thread %d: computes C[%d][%d]\n", threadID, row, col)
	sum := 0
	for p := 0; p < size; p++ {
		sum += A[row][p] * B[p][col]
	}
	C[row][col] = sum
}

func workConsecutiveRow(A, B, C Matrix, size, startIdx, endIdx, threadID int, wg *sync.WaitGroup) {
	defer wg.Done()

	for I := startIdx; I < endIdx; I++ {
		row := I / size
		col := I % size

		computeElement(A, B, C, row, col, size, threadID)
	}
}

func workConsecutiveCol(A, B, C Matrix, size, startIdx, endIdx, threadID int, wg *sync.WaitGroup) {
	defer wg.Done()

	for I := startIdx; I < endIdx; I++ {
		col := I / size
		row := I % size

		computeElement(A, B, C, row, col, size, threadID)
	}
}

func workInterleavedRow(A, B, C Matrix, size, startIdx, nrThreads, threadID int, wg *sync.WaitGroup) {
	defer wg.Done()
	totalElements := size * size

	for I := startIdx; I < totalElements; I += nrThreads {
		row := I / size
		col := I % size

		computeElement(A, B, C, row, col, size, threadID)
	}
}

type workStrategy func(A, B, C Matrix, size, startIdx, endIdx, threadID int, wg *sync.WaitGroup)

func parallelMultiplyManager(A, B, C Matrix, size, numThreads int, strategy workStrategy) {
	totalElements := size * size
	baseWork := totalElements / numThreads
	remainder := totalElements % numThreads

	var wg sync.WaitGroup
	currentStartIdx := 0

	for i := 0; i < numThreads; i++ {
		workSize := baseWork
		if i < remainder {
			workSize++
		}

		endIdx := currentStartIdx + workSize

		wg.Add(1)
		go strategy(A, B, C, size, currentStartIdx, endIdx, i, &wg)
		currentStartIdx = endIdx
	}

	wg.Wait()
}

func parallelMultiplyInterleavedManager(A, B, C Matrix, size, numThreads int) {
	var wg sync.WaitGroup

	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go workInterleavedRow(A, B, C, size, i, numThreads, i, &wg)
	}

	wg.Wait()
}

func newMatrix(size int) Matrix {
	matrix := make(Matrix, size)
	for i := 0; i < size; i++ {
		matrix[i] = make([]int, size)
		for j := 0; j < size; j++ {
			matrix[i][j] = rand.Intn(MaxVal)
		}
	}
	return matrix
}

func zeroMatrix(size int) Matrix {
	matrix := make(Matrix, size)
	for i := 0; i < size; i++ {
		matrix[i] = make([]int, size)
	}
	return matrix
}

func main() {

	matrixSize := 3000
	nrThreads := 100

	rand.Seed(time.Now().UnixNano())
	A := newMatrix(matrixSize)
	B := newMatrix(matrixSize)

	C1 := zeroMatrix(matrixSize)
	start := time.Now()
	parallelMultiplyManager(A, B, C1, matrixSize, nrThreads, workConsecutiveRow)
	elapsed := time.Since(start)
	fmt.Printf("Strategy 1 (Consecutive Row-Major): %v\n", elapsed)

	C2 := zeroMatrix(matrixSize)
	start = time.Now()
	parallelMultiplyManager(A, B, C2, matrixSize, nrThreads, workConsecutiveCol)
	elapsed = time.Since(start)
	fmt.Printf("Strategy 2 (Consecutive Col-Major): %v\n", elapsed)

	C3 := zeroMatrix(matrixSize)
	start = time.Now()
	parallelMultiplyInterleavedManager(A, B, C3, matrixSize, nrThreads)
	elapsed = time.Since(start)
	fmt.Printf("Strategy 3 (Interleaved Row-Major): %v\n", elapsed)
}
