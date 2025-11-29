package main

import (
	"fmt"
	"lab6-go/graph"
	"lab6-go/solver"
	"log"
	"math/rand"
	"time"
)

func main() {

	rand.New(rand.NewSource(time.Now().UnixNano()))
	numThreads := 10
	startVertex := 0
	nrVertices := 1000
	nrEdges := nrVertices * 2

	g := graph.NewRandomGraph(nrVertices, nrEdges)
	//g, err := graph.NewGraphFromFile("graph2.txt")
	fmt.Println("Graph created")

	s1 := solver.NewReusableThreadsSolver(g)
	start := time.Now()
	solution, err := s1.FindCycle(numThreads, startVertex)
	elapsed := time.Since(start)
	fmt.Printf("Time for optimized one is: %v\n", elapsed)

	s2 := solver.NewSolver(g)
	start = time.Now()
	solution, err = s2.FindCycle(numThreads, startVertex)
	elapsed = time.Since(start)
	fmt.Printf("Time for non reusable one is: %v\n", elapsed)

	s3 := solver.NewSolver(g)
	start = time.Now()
	solution, err = s3.FindCycle(1, startVertex)
	elapsed = time.Since(start)
	fmt.Printf("Time for no parallel one is: %v\n", elapsed)

	if err != nil {
		log.Printf("Search complete: %v\n", err)
	} else {
		fmt.Println(solution)
		// _ = g.SaveToFile("graph3.txt")
	}
}
