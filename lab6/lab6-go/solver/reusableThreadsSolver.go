package solver

import (
	"fmt"
	"lab6-go/graph"
	"sync"
	"sync/atomic"
)

type ReusableThreadsSolver struct {
	graph *graph.Graph

	solution      []int
	solutionLock  sync.Mutex
	solutionFound int32
}

func NewReusableThreadsSolver(g *graph.Graph) *ReusableThreadsSolver {
	return &ReusableThreadsSolver{
		graph: g,
	}
}

func (s *ReusableThreadsSolver) FindCycle(nrThreads, startVertex int) ([]int, error) {
	atomic.StoreInt32(&s.solutionFound, 0)
	s.solution = nil

	sem := make(chan struct{}, nrThreads)
	sem <- struct{}{}
	defer func() { <-sem }()

	path := make([]int, 1, s.graph.NumVertices())
	path[0] = startVertex
	visited := make([]bool, s.graph.NumVertices())
	visited[startVertex] = true

	s.recursiveSearch(path, visited, sem, startVertex)

	if atomic.LoadInt32(&s.solutionFound) == 1 {
		s.solutionLock.Lock()
		defer s.solutionLock.Unlock()
		solCopy := make([]int, len(s.solution))
		copy(solCopy, s.solution)
		return solCopy, nil
	}

	return nil, fmt.Errorf("no Hamiltonian cycle found")
}

func (s *ReusableThreadsSolver) recursiveSearch(
	path []int,
	visited []bool,
	sem chan struct{},
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

	s.searchParallel(path, visited, sem, validNeighbors)
}

func (s *ReusableThreadsSolver) searchParallel(path []int, visited []bool, sem chan struct{}, validNeighbors []int) {

	var wg sync.WaitGroup
	wg.Add(len(validNeighbors))

	for _, next := range validNeighbors {
		if atomic.LoadInt32(&s.solutionFound) == 1 {
			wg.Done()
			continue
		}

		select {
		case sem <- struct{}{}:
			newPath, newVisited := cloneState(path, visited, next, s.graph.NumVertices())

			go func(v int, p []int, vis []bool) {
				defer wg.Done()
				s.recursiveSearch(p, vis, sem, v)
				<-sem
			}(next, newPath, newVisited)

		default:
			if atomic.LoadInt32(&s.solutionFound) == 1 {
				wg.Done()
				continue
			}

			visited[next] = true
			path = append(path, next)

			s.recursiveSearch(path, visited, sem, next)

			path = path[:len(path)-1]
			visited[next] = false
			wg.Done()
		}
	}
	wg.Wait()
}
