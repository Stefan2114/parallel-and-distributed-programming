# Parallel Polynomial Multiplication: MPI vs. Multi-threading

This project explores parallel polynomial multiplication using two distinct approaches:
1.  **Lab 5:** Shared Memory Concurrency (Go Goroutines & Channels).
2.  **Lab 7:** Distributed Memory Parallelism (MPI using `gompi`).

## 1. Algorithms

We compared two multiplication algorithms:

### Regular Multiplication ($O(n^2)$)
* **Logic:** Iterates through every element of the first polynomial and multiplies it by every element of the second.
* **Complexity:** $O(n^2)$.
* **Pros:** Simple to implement; highly parallelizable (embarrassingly parallel).
* **Cons:** Becomes extremely slow as input size $N$ grows.

### Karatsuba Multiplication ($O(n^{\log_2 3})$)
* **Logic:** A divide-and-conquer algorithm that splits polynomials into high and low halves to reduce the number of multiplications from 4 to 3 at each recursive step.
* **Complexity:** $\approx O(n^{1.585})$.
* **Pros:** Significantly faster for large $N$.

---

## 2. Implementation & Distribution Strategies

### Lab 5: Shared Memory (Goroutines)
* **O(n^2) Parallel:** Utilized a worker pool (Fixed Threads) where the result array is split into chunks. Each goroutine computes a specific range of indices.
* **Karatsuba Parallel:** Used `sync.WaitGroup` and limited recursion depth (or semaphores) to spawn goroutines for the 3 recursive calls ($P_{high}Q_{high}$, $P_{low}Q_{low}$, and middle terms) in parallel.

### Lab 7: Distributed Memory (MPI)
* **Communication:** Used `gob` serialization to send `[]*big.Int` slices over MPI channels.
* **O(n^2) Distributed:**
    * **Coordinator (Rank 0):** Broadcasts both full polynomials to all workers. Assigns a specific range of the *result* indices to each worker.
    * **Workers:** Compute their chunk and send the sub-result back to Rank 0 to be gathered.
* **Karatsuba Distributed (Task Parallelism):**
    * Designed for 4 processes (1 Master + 3 Workers).
    * **Rank 0:** Splits polynomials into High/Low parts and sums.
    * **Rank 1:** Receives High parts, calculates $P_{high} \times Q_{high}$.
    * **Rank 2:** Receives Low parts, calculates $P_{low} \times Q_{low}$.
    * **Rank 3:** Receives Sum parts, calculates $(P_{high}+P_{low}) \times (Q_{high}+Q_{low})$.
    * **Rank 0:** Gathers results and applies the Karatsuba combination formula.

---

## 3. Performance Analysis

The following tests compare execution time on the same hardware.

### Dataset: Size 10,000

| Method    | Algorithm | Implementation | Time |
|:----------| :--- | :--- | :--- |
| **Lab 5** | $O(n^2)$ | Go Parallel (16 threads) | ~452ms |
| **Lab 7** | $O(n^2)$ | MPI Distributed (4 procs) | **~412ms** |
| **Lab 5** | Karatsuba | Go Parallel (16 threads) | **~133ms** |
| **Lab 7** | Karatsuba | MPI Distributed (4 procs) | ~152ms |

### Dataset: Size 100,000

| Method    | Algorithm | Implementation | Time |
|:----------| :--- | :--- | :--- |
| **Lab 5** | $O(n^2)$ | Go Parallel (16 threads) | **43.86s** |
| **Lab 7** | $O(n^2)$ | MPI Distributed (4 procs) | **45.32s** |
| **Lab 5** | Karatsuba | Go Parallel (16 threads) | **4.23s** |
| **Lab 7** | Karatsuba | MPI Distributed (4 procs) | **5.13s** |

### Observations & Conclusions

1.  **MPI vs Shared Memory Competitiveness:**
    Contrary to initial assumptions about serialization overhead, the MPI implementation is **highly competitive** with the Shared Memory implementation.
    * For the small dataset (10,000), MPI was actually **faster** for the $O(n^2)$ algorithm (412ms vs 452ms).
    * For the large dataset (100,000), MPI was only marginally slower (approx 1-3% difference in Regular Multiplication).
    * This suggests that the overhead of `gob` serialization and IPC (Inter-Process Communication) is well-managed and balanced by the efficient separation of memory spaces.

2.  **Karatsuba Efficiency:**
    Regardless of the parallelization method, Karatsuba offers massive speedups over $O(n^2)$ for large inputs. In Lab 5, Karatsuba provided a **~10x speedup** (4.23s vs 43.86s) for the 100k dataset.

3.  **Scalability Implications:**
    Since the MPI implementation performs nearly identically to the Shared Memory implementation on a single machine, it is the superior architectural choice for larger scales. Lab 5 is physically limited to the cores of one CPU, whereas Lab 7 can be scaled horizontally across a cluster of machines to handle datasets far larger than 100,000 elements.

## 4. Optimization: Data Type Impact (int64 vs big.Int)

To isolate the cost of arithmetic operations versus parallel overhead, we ran a separate set of benchmarks using fixed-precision integers (`int64`) instead of arbitrary-precision integers (`big.Int`).

### int64 Performance Results

**Dataset: Size 10,000**

| Method    | Algorithm | Implementation | Time |
|:----------| :--- | :--- | :--- |
| **Lab 5** | $O(n^2)$ | Go Parallel (16 threads) | **21.12ms** |
| **Lab 7** | $O(n^2)$ | MPI Distributed (4 procs) | **24.64ms** |
| **Lab 5** | Karatsuba | Go Parallel (16 threads) | **9.05ms** |
| **Lab 7** | Karatsuba | MPI Distributed (4 procs) | **15.82ms** |

**Dataset: Size 100,000**

| Method    | Algorithm | Implementation | Time |
|:----------| :--- | :--- | :--- |
| **Lab 5** | $O(n^2)$ | Go Parallel (16 threads) | **1.78s** |
| **Lab 7** | $O(n^2)$ | MPI Distributed (4 procs) | **1.92s** |
| **Lab 5** | Karatsuba | Go Parallel (16 threads) | **129ms** |
| **Lab 7** | Karatsuba | MPI Distributed (4 procs) | **175ms** |

### Comparative Analysis: int64 vs big.Int

1.  **Massive Speedup (~25x - 30x):**
    Switching to `int64` resulted in a drastic performance improvement.
    * For $O(n^2)$ 100k: Time dropped from **~44s** (`big.Int`) to **~1.78s** (`int64`).
    * For Karatsuba 100k: Time dropped from **~4.2s** (`big.Int`) to **~129ms** (`int64`).
    * **Reason:** `big.Int` requires dynamic memory allocation for every number and software-implemented arithmetic. `int64` uses native CPU registers and instructions, removing the allocation overhead and significantly speeding up the math.

2.  **Impact on MPI Overhead:**
    With `big.Int`, the computation time was so high that it masked some of the MPI communication costs. With `int64`, the calculation is lightning fast, making the communication overhead (latency and serialization) more visible.
    * *Example (Karatsuba 100k):* MPI is ~35% slower than Shared Memory (175ms vs 129ms).
    * However, the fact that MPI remains within milliseconds of the Shared Memory implementation proves that the `gob` serialization for `int64` is extremely efficient compared to `big.Int` structures.