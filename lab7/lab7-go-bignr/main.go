package main

import (
	"fmt"
	"math/big"
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

	var p1, p2 []*big.Int

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
				if resDistSimple[i].Cmp(resDistKara[i]) != 0 {
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
//Time for O(n^2) with nr threads: 16 is: 43.863720505s
//Time for O(n^2) in parallel is: 44.418321261s
//Time for Karatsuba (n^log_2(3)) is: 17.439350191s
//Time for Karatsuba with nr threads: 16 is: 4.225685506s
//Time for Karatsuba with 3^n threads is: 4.654790385s

//// With MPI
//[O(n^2) Distributed] Time with 4 processes: 45.324080013s
//[Karatsuba Distributed] Time with 4 processes: 5.125144594s

// 10000
//Time for O(n^2) with nr threads: 16 is: 452.400304ms
//Time for O(n^2) in parallel is: 351.59677ms
//Time for Karatsuba (n^log_2(3)) is: 546.071045ms
//Time for Karatsuba with nr threads: 16 is: 133.077587ms
//Time for Karatsuba with 3^n threads is: 129.26906ms

// With MPI
//[O(n^2) Distributed] Time with 4 processes: 412.211327ms
//[Karatsuba Distributed] Time with 4 processes: 152.08227ms
