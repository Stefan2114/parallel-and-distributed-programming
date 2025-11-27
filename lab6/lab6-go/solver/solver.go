package solver

import (
	"fmt"
	"lab6-go/graph"
	"sync"
	"sync/atomic"
)

type Solver struct {
	graph *graph.Graph

	solution      []int
	solutionLock  sync.Mutex
	solutionFound int32
}

func NewSolver(g *graph.Graph) *Solver {
	return &Solver{
		graph: g,
	}
}

func (s *Solver) FindCycle(nrThreads, startVertex int) ([]int, error) {
	// Reset state for new search
	atomic.StoreInt32(&s.solutionFound, 0)
	s.solution = nil

	path := make([]int, 1, s.graph.NumVertices())
	path[0] = startVertex
	visited := make([]bool, s.graph.NumVertices())
	visited[startVertex] = true

	s.recursiveSearch(path, visited, nrThreads, startVertex)

	if atomic.LoadInt32(&s.solutionFound) == 1 {
		s.solutionLock.Lock()
		defer s.solutionLock.Unlock()
		solCopy := make([]int, len(s.solution))
		copy(solCopy, s.solution)
		return solCopy, nil
	}

	return nil, fmt.Errorf("no Hamiltonian cycle found")
}

func (s *Solver) recursiveSearch(
	path []int,
	visited []bool,
	nrThreads int,
	currentVertex int,
) {

	if atomic.LoadInt32(&s.solutionFound) == 1 {
		return
	}

	if len(path) == s.graph.NumVertices() {
		startVertex := path[0]
		if s.graph.HasEdge(currentVertex, startVertex) {
			if atomic.CompareAndSwapInt32(&s.solutionFound, 0, 1) {
				s.solutionLock.Lock()
				s.solution = make([]int, len(path))
				copy(s.solution, path)
				s.solutionLock.Unlock()
			}
		}
		return
	}

	validNeighbors := getValidNeighbors(s.graph.GetNeighbors(currentVertex), visited)
	if len(validNeighbors) == 0 {
		return
	}

	if nrThreads == 1 {
		s.searchSequential(path, visited, validNeighbors)
	} else {
		s.searchParallel(path, visited, nrThreads, validNeighbors)
	}
}

func (s *Solver) searchSequential(path []int, visited []bool, validNeighbors []int) {
	for _, next := range validNeighbors {
		if atomic.LoadInt32(&s.solutionFound) == 1 {
			return
		}

		visited[next] = true
		path = append(path, next)
		s.recursiveSearch(path, visited, 1, next)

		path = path[:len(path)-1]
		visited[next] = false
	}
}

func (s *Solver) searchParallel(path []int, visited []bool, nrThreads int, validNeighbors []int) {
	nrBranches := len(validNeighbors)
	baseThreads := nrThreads / nrBranches
	extraThreads := nrThreads % nrBranches
	var wg sync.WaitGroup

	for i, next := range validNeighbors {
		if atomic.LoadInt32(&s.solutionFound) == 1 {
			break
		}

		threadsPerBranch := baseThreads
		if i < extraThreads {
			threadsPerBranch++
		}

		newPath, newVisited := cloneState(path, visited, next, s.graph.NumVertices())

		if threadsPerBranch > 0 {
			wg.Add(1)
			go func(v int, t int, p []int, vis []bool) {
				defer wg.Done()
				s.recursiveSearch(p, vis, t, v)
			}(next, threadsPerBranch, newPath, newVisited)
		} else {
			s.recursiveSearch(newPath, newVisited, 1, next)
		}
	}
	wg.Wait()
}

func getValidNeighbors(neighbors []int, visited []bool) []int {
	validNeighbors := make([]int, 0, len(neighbors))
	for _, neighbor := range neighbors {
		if !visited[neighbor] {
			validNeighbors = append(validNeighbors, neighbor)
		}
	}
	return validNeighbors
}

func cloneState(path []int, visited []bool, nextVertex int, capacity int) ([]int, []bool) {
	newPath := make([]int, len(path), capacity)
	copy(newPath, path)
	newPath = append(newPath, nextVertex)

	newVisited := make([]bool, len(visited))
	copy(newVisited, visited)
	newVisited[nextVertex] = true

	return newPath, newVisited
}
