# Lab 6: Parallel Hamiltonian Cycle Search

## 1. Algorithms Description

The solution implements a **Backtracking Search** to find a Hamiltonian cycle (a path visiting every vertex exactly once and returning to the start). To optimize performance, the search is parallelized by splitting the exploration of neighbor nodes across available threads.

### Parallelization Strategy
The core logic distributes work based on a "thread budget":
1.  **Thread Allocation:** Given $N$ threads and a vertex with $K$ valid neighbors, the threads are divided among branches ($N/K$).
2.  **Branching:** * If allocated threads > 1: A new asynchronous task (Goroutine or ForkJoinTask) is spawned for that neighbor.
    * If allocated threads = 1: The search continues sequentially on the current thread to avoid context-switching overhead.
3.  **State Management:** Because parallel branches run simultaneously, each branch works on a **deep copy** of the `path` and `visited` array to prevent race conditions.

### Implementations
* **Go (Standard):** Uses `sync.WaitGroup` to spawn Goroutines based strictly on the split calculation.
* **Go (Reusable/Optimized):** Uses a **Semaphore** (buffered channel) to limit active Goroutines. If the semaphore is full, the current thread executes the branch sequentially instead of spawning a new one.
* **Java:** Uses `ForkJoinPool` and `RecursiveTask`. The framework handles work-stealing and task management automatically.

---

## 2. Synchronization

To ensure correctness in a multi-threaded environment, the following synchronization mechanisms were used:

* **Solution Found Flag:** An atomic boolean (`atomic.Int32` in Go, `AtomicBoolean` in Java) is used to signal all running threads to stop immediately once a valid cycle is found.
* **Solution Storage:** A `Mutex` (Go) protects the writing of the final path to ensure the data is consistent before returning.
* **Wait Groups:** In Go, `sync.WaitGroup` ensures the parent thread waits for all parallel child branches to complete (or abort) before returning.
* **Java ForkJoin:** Uses `.fork()` and `.join()` which internally handles synchronization of child tasks.

---

## 3. Go Implementation Analysis: Reusable vs. Standard

I implemented two versions in Go to test thread management strategies.

* **The Optimization:** The `ReusableThreadsSolver` uses a semaphore to cap the number of active Goroutines. Instead of blindly spawning threads, it checks if a "slot" is open. If not, it reuses the current thread.
* **The Trade-off:**
    * **Pros:** Reduces the overhead of the Go runtime scheduler when thousands of branches exist. It performs significantly better on **Graph 2**, cutting the time by more than half (3.24s vs 7.02s).
    * **Cons:** Both parallel approaches require **deep copying** the `path` and `visited` arrays for every parallel branch. In graphs where the search depth is shallow or finding neighbors is fast (like **Graph 3**), the memory overhead of copying state outweighs the CPU gain, making the "optimized" version slower than the standard one.

---

## 4. Performance Measurements

The following tests were run on 3 random directed graphs. The sequential solution was often too slow to complete within reasonable time limits.

| Scenario | Go (Standard) | Go (Optimized/Reusable) | Java (ForkJoin) | Sequential |
| :--- | :--- | :--- | :--- | :--- |
| **Graph 1** (38v, 152e, 16t) | **0.12 s** | 0.15 s | 0.44 s | 20 s |
| **Graph 2** (50v, 250e, 16t) | 7.02 s | 3.24 s | **3.1 s** | > 2 min (Time Limit) |
| **Graph 3** (40v, 200e, 16t) | 4.90 s | 13.36 s | **2.26 s** | > 2 min (Time Limit) |
| **Graph 1** (38v, 152e, 100t) | **0.54 s** | 0.87 s | 0.56 s | 20 s |
| **Graph 2** (50v, 250e, 100t) | 16.4 s | **0.6 s** | 14.87 s | > 2 min (Time Limit) |
| **Graph 3** (40v, 200e, 100t) | 15.51 s | **3.3 s** | 6.22 s | > 2 min (Time Limit) |