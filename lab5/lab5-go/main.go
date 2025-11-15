package main

import (
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"
)

const MaxVal = 30000000

func computeElement(k int, polyA, polyB, result []*big.Int) {

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
	result[k] = localSum
}

func PolyMulParallel(polyA, polyB []*big.Int) []*big.Int {

	lenA := len(polyA)
	lenB := len(polyB)
	if lenA == 0 || lenB == 0 {
		return []*big.Int{}
	}

	resultLen := lenA + lenB - 1
	result := make([]*big.Int, resultLen)
	var wg sync.WaitGroup

	for k := 0; k < resultLen; k++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			computeElement(k, polyA, polyB, result)
		}(k)
	}

	wg.Wait()
	return result
}

func PolyMulParallelWithFixNrThreads(polyA, polyB []*big.Int, nrThreads int) []*big.Int {

	lenA := len(polyA)
	lenB := len(polyB)
	if lenA == 0 || lenB == 0 {
		return []*big.Int{}
	}

	resultLen := lenA + lenB - 1
	result := make([]*big.Int, resultLen)
	var wg sync.WaitGroup
	baseWork := resultLen / nrThreads
	remainder := resultLen % nrThreads
	currentStartIdx := 0

	for k := 0; k < nrThreads; k++ {

		workSize := baseWork
		if k < remainder {
			workSize++
		}

		endIdx := currentStartIdx + workSize

		wg.Add(1)
		go func(startIndex, endIndex int) {
			defer wg.Done()
			for i := startIndex; i < endIndex; i++ {
				computeElement(i, polyA, polyB, result)
			}
		}(currentStartIdx, endIdx)
		currentStartIdx = endIdx
	}

	wg.Wait()
	return result
}

func newPolynomial(size int) []*big.Int {
	polynomial := make([]*big.Int, size)
	for i := 0; i < size; i++ {
		polynomial[i] = big.NewInt(rand.Int63n(MaxVal))
	}
	return polynomial
}

func arePolynomialsEqual(p1, p2 []*big.Int) bool {

	if len(p1) != len(p2) {
		return false
	}
	for i := 0; i < len(p1); i++ {
		if p1[i].Cmp(p2[i]) != 0 {
			return false
		}
	}
	return true
}

func main() {

	size1 := 100000
	size2 := 100000

	nrThreads := 16

	rand.New(rand.NewSource(time.Now().UnixNano()))
	p1 := newPolynomial(size1)
	p2 := newPolynomial(size2)

	start := time.Now()
	result1 := PolyMulSequential(p1, p2)
	elapsed := time.Since(start)
	fmt.Printf("Time for O(n^2) swquential is: %v\n", elapsed)

	start = time.Now()
	result2 := PolyMulParallelWithFixNrThreads(p1, p2, nrThreads)
	elapsed = time.Since(start)
	fmt.Printf("Time for O(n^2) with nr threads: %d is: %v\n", nrThreads, elapsed)

	if !arePolynomialsEqual(result1, result2) {
		panic("Not not equal")
	}

	start = time.Now()
	result3 := PolyMulParallel(p1, p2)
	elapsed = time.Since(start)
	fmt.Printf("Time for O(n^2) in parallel is: %v\n", elapsed)

	if !arePolynomialsEqual(result1, result3) {
		panic("Not not equal")
	}

	start = time.Now()
	result4 := PolyMulKaratsuba(p1, p2)
	elapsed = time.Since(start)
	fmt.Printf("Time for Karatsuba (n^log_2(3)) is: %v\n", elapsed)

	if !arePolynomialsEqual(result1, result4) {
		panic("Not equal")
	}

	start = time.Now()
	result5 := polyMulKaratsubaParallel(p1, p2, nrThreads)
	elapsed = time.Since(start)
	fmt.Printf("Time for Karatsuba with nr threads: %d is: %v\n", nrThreads, elapsed)

	if !arePolynomialsEqual(result1, result5) {
		panic("Not equal")
	}

	start = time.Now()
	result6 := PolyMulKaratsubaParallelFine(p1, p2)
	elapsed = time.Since(start)
	fmt.Printf("Time for Karatsuba with 3^n threads is: %v\n", elapsed)

	if !arePolynomialsEqual(result1, result6) {
		panic("Not equal")
	}
}
