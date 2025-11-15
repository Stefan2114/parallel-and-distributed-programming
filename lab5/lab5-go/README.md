# Lab 5: Parallel Polynomial Multiplication

---

## 1. Algorithms

* **O(n²) Algorithm:** The standard "schoolbook" method. Implemented as `PolyMulSequential`. It computes each coefficient $k$ of the result by summing all products $P_i \times Q_j$ where $i+j=k$.

* **Karatsuba Algorithm:** A recursive $O(n^{\log_2 3})$ "divide and conquer" algorithm. Implemented as `PolyMulKaratsuba`.
    * **Mathematical Logic:** It splits each polynomial into a high and low part:
        * $P(X) = P_1(X) \cdot X^n + P_2(X)$
        * $Q(X) = Q_1(X) \cdot X^n + Q_2(X)$
    * A naive multiply would require 4 recursive calls ($P_1 Q_1$, $P_1 Q_2$, $P_2 Q_1$, $P_2 Q_2$).
    * Karatsuba's "trick" is to compute the middle term using only one new multiplication, for a total of 3:
        1.  **$R_{\text{high}}$** = $P_1 \times Q_1$
        2.  **$R_{\text{low}}$** = $P_2 \times Q_2$
        3.  **$R_{\text{mid\_term}}$** = $(P_1+P_2) \times (Q_1+Q_2)$
    * The final middle term is found with fast subtraction:
        * $R_{\text{mid}} = R_{\text{mid\_term}} - R_{\text{high}} - R_{\text{low}}$
    * The final result is combined:
        * $\text{Result} = (R_{\text{high}} \cdot X^{2n}) + (R_{\text{mid}} \cdot X^n) + R_{\text{low}}$

---

## 2. Synchronization

* **O(n²) Parallel (Fixed - `PolyMulParallelWithFixNrThreads`):**
    * Uses a **`sync.WaitGroup`**.
    * Divides the work into `nrThreads` chunks. One goroutine is launched per chunk, and `wg.Wait()` waits for all to finish.

* **O(n²) Parallel (Unbounded - `PolyMulParallel`):**
    * Uses a **`sync.WaitGroup`**.
    * Launches **one goroutine for every coefficient** in the result. Inefficient for large polynomials due to scheduler thrashing.

* **Karatsuba Parallel (Hybrid - `polyMulKaratsubaParallel`):**
    * Uses a **hybrid "try-acquire" pattern**.
    * A **`sync.WaitGroup`** waits for the 3 recursive calls.
    * A **buffered channel (`chan struct{}`)** is used as a **semaphore** to limit active goroutines to `nrThreads`.
    * A **`select`/`default`** statement tries to acquire a slot. If successful, it spawns a goroutine. If not (pool is full), it runs the recursive call *sequentially*.

* **Karatsuba Parallel (Fine - `PolyMulKaratsubaParallelFine`):**
    * Uses a **`sync.WaitGroup`** at *each recursive step*.
    * Spawns 2 new goroutines at each step, creating a $3^k$ "fork-bomb" of goroutines.

---

## 3. Performance Measurements

Measurements for `n=10,000` on a multi-core CPU.

**Run: `nrThreads = 16`**

| Algorithm | Variant                   | Time       |
| :--- |:--------------------------|:-----------|
| O(n²) | Sequential                | 2.67 s     |
| O(n²) | Parallel (Fixed 16)       | 416 ms     |
| O(n²) | Parallel (Unbounded)      | 834 ms     |
| Karatsuba | Sequential                | 627 ms     |
| **Karatsuba** | Parallel (Hybrid 16)      | 218 ms     |
| Karatsuba | **Parallel (Fine "3^k")** | **144 ms** |

---

Measurements for `n=100,000` on a multi-core CPU.

**Run: `nrThreads = 16`**

| Algorithm | Variant                  | Time      |
| :--- |:-------------------------|:----------|
| O(n²) | Sequential               | 3.38 m    |
| O(n²) | Parallel (Fixed 16)      | 46.13 s   |
| O(n²) | Parallel (Unbounded)     | 41.85 s   |
| Karatsuba | Sequential               | 18.32 s   |
| **Karatsuba** | **Parallel (Hybrid 16)** | **4.34 s** |
| Karatsuba | Parallel (Fine "3^k")    | 5.03 s    |

### Analysis

1.  **O(n²) vs. Karatsuba:** The sequential Karatsuba algorithm proves its superior complexity. It's **~4.3x faster** than sequential O(n²) at n=10,000 (627 ms vs 2.67 s). This advantage widens significantly at n=100,000, where Karatsuba is **~11.1x faster** (18.32 s vs 3.38 m).

2.  **O(n²) Parallel:** Parallelism provides a major speedup, but the best approach is inconsistent. The "Fixed 16" variant was fastest at n=10,000 (416 ms), while the "Unbounded" variant was fastest at n=100,000 (41.85 s).

3.  **Karatsuba Parallel:**
    * At n=10,000, the **Parallel (Fine "3^k")** was the fastest overall (144 ms), achieving a 4.35x speedup over sequential Karatsuba.
    * At n=100,000, the **Parallel (Hybrid 16)** was the fastest overall (4.34 s), achieving a 4.22x speedup over sequential Karatsuba.

4.  **Effect of `nrThreads`:`nrThreads` was set to an excessive 2,000,000. This *slowed down* both the "Fixed" and "Hybrid" parallel versions (to 2.4s and 294ms, respectively) due to the overhead of managing such a large (and unnecessary) thread/semaphore pool.

### Conclusion

The **Karatsuba** algorithm is the clear algorithmic winner, and its performance gap over O(n²) grows dramatically with the problem size.
While the "Fine" parallel model was fastest on the smaller dataset, the Parallel (Hybrid 16) Karatsuba is the best and most scalable solution, as it performed best on the largest, most complex workload (n=100,000). This shows that its controlled, throttled approach to parallelism is more efficient and robust for larger tasks than the "fork-bomb" (Fine) model.