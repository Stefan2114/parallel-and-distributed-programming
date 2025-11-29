package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	MpiInit()
	defer MpiFinalize()

	rank := MpiRank()
	worldSize := MpiSize()

	// Settings
	size1 := 100000
	size2 := 100000

	var p1, p2 []int64

	if rank == 0 {
		rand.New(rand.NewSource(time.Now().UnixNano()))
		p1 = newPolynomial(size1)
		p2 = newPolynomial(size2)
		fmt.Printf("Rank 0: Generated polynomials size %d and %d\n", size1, size2)
	}

	// 1. Run Distributed O(n^2)
	MpiBarrier()
	start := time.Now()

	resDistSimple := PolyMulDistributedSimple(p1, p2, rank, worldSize)

	MpiBarrier()
	elapsed := time.Since(start)

	if rank == 0 {
		fmt.Printf("[O(n^2) Distributed] Time with %d processes: %v\n", worldSize, elapsed)
	}

	// 2. Run Distributed Karatsuba
	MpiBarrier()
	start = time.Now()

	resDistKara := PolyMulDistributedKaratsuba(p1, p2, rank, worldSize)

	MpiBarrier()
	elapsed = time.Since(start)

	if rank == 0 {
		fmt.Printf("[Karatsuba Distributed] Time with %d processes: %v\n", worldSize, elapsed)

		if len(resDistSimple) != len(resDistKara) {
			fmt.Printf("ERROR: Length mismatch! Simple: %d, Kara: %d\n", len(resDistSimple), len(resDistKara))
		} else {
			match := true
			for i := range resDistSimple {
				if resDistSimple[i]-resDistKara[i] != 0 {
					match = false
					break
				}
			}
			if !match {
				fmt.Println("ERROR: Results mismatch.")
			}
		}
	}
}

// 100000
//Time for O(n^2) with nr threads: 16 is: 1.782710761s
//Time for Karatsuba with nr threads: 16 is: 129.393755ms

//// With MPI
//[O(n^2) Distributed] Time with 4 processes: 1.916132719s
//[Karatsuba Distributed] Time with 4 processes: 174.533896ms

// 10000
//Time for O(n^2) with nr threads: 16 is: 21.122182ms
//Time for Karatsuba with nr threads: 16 is: 9.051721ms

// With MPI
//[O(n^2) Distributed] Time with 4 processes: 24.644807ms
//[Karatsuba Distributed] Time with 4 processes: 15.815475ms
