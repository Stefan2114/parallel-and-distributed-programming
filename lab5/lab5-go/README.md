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

Measurements for `n=100,000` on a multi-core CPU.

**Run 1: `nrThreads = 1000`**

| Algorithm | Variant | Time |
| :--- | :--- | :--- |
| O(n²) | Sequential | 9.059 s |
| O(n²) | Parallel (Fixed 1000) | 2.066 s |
| O(n²) | Parallel (Unbounded) | 1.836 s |
| Karatsuba | Sequential | 612 ms |
| **Karatsuba** | **Parallel (Hybrid 1000)** | **206 ms** |
| Karatsuba | Parallel (Fine "3^k") | 274 ms |

### Analysis

1.  **O(n²) vs. Karatsuba:** The sequential Karatsuba (612 ms) is **~15x faster** than the sequential O(n²) (9059 ms), proving its superior algorithmic complexity for this problem size.

2.  **O(n²) Parallel:** Both parallel O(n²) versions are significantly faster than the sequential one. The "Unbounded" version (1.8s) was slightly faster than the "Fixed" (2.0s), suggesting the Go scheduler handled the massive number of goroutines better than our manual chunking.

3.  **Karatsuba Parallel:**
    * The **Hybrid (Hybrid 1000)** version is the **fastest overall (206 ms)**, showing a 3x speedup over sequential Karatsuba.
    * The **Fine "3^k"** (274 ms) was also significantly faster than sequential Karatsuba, but slower than the Hybrid. This shows the Go scheduler is surprisingly good at managing the "fork-bomb," but the controlled semaphore (Hybrid) approach is still more efficient.

4.  **Effect of `nrThreads`:`nrThreads` was set to an excessive 2,000,000. This *slowed down* both the "Fixed" and "Hybrid" parallel versions (to 2.4s and 294ms, respectively) due to the overhead of managing such a large (and unnecessary) thread/semaphore pool.

### Conclusion

The **Hybrid Parallel Karatsuba** algorithm (`polyMulKaratsubaParallel`) combined with a right amount of threads is the clear winner, as it uses the fastest algorithm ($O(n^{1.58})$) with an efficient, throttled parallel model. A moderate `nrThreads` (e.g., 1000) provides the best performance, while an excessively large value adds overhead and slows the program down.