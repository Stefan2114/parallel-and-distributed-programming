package graph

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
)

type Graph struct {
	nrVertices int
	adjList    map[int][]int
}

func NewGraph(nrVertices int) *Graph {
	return &Graph{nrVertices: nrVertices, adjList: make(map[int][]int)}
}
func (g *Graph) AddEdge(from, to int) {
	g.adjList[from] = append(g.adjList[from], to)
}
func (g *Graph) GetNeighbors(vertex int) []int {
	if neighbors, ok := g.adjList[vertex]; ok {
		return neighbors
	}
	return []int{}
}

func (g *Graph) HasEdge(from, to int) bool {
	for _, neighbor := range g.GetNeighbors(from) {
		if neighbor == to {
			return true
		}
	}
	return false
}
func (g *Graph) NumVertices() int {
	return g.nrVertices
}

func NewRandomGraph(nrVertices, edges int) *Graph {
	g := &Graph{nrVertices: nrVertices, adjList: make(map[int][]int)}
	for i := 0; i < nrVertices; i++ {
		for j := 0; j < nrVertices; j++ {
			if i != j {
				nr := rand.IntN(10)
				if nr < 2 {
					g.AddEdge(i, j)
				}
			}
		}

	}
	return g
}

func (g *Graph) edgeAlreadyExists(from, to int) bool {
	for _, v := range g.adjList[from] {
		if v == to {
			return true
		}
	}
	return false
}

func (g *Graph) SaveToFile(filename string) error {

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	_, err = fmt.Fprintln(writer, g.nrVertices)
	if err != nil {
		return err
	}

	for from := 0; from < g.nrVertices; from++ {
		neighbors, exists := g.adjList[from]
		if exists {
			for _, to := range neighbors {
				_, err := fmt.Fprintln(writer, from, to)
				if err != nil {
					return err
				}
			}
		}
	}

	return writer.Flush()
}

func NewGraphFromFile(filename string) (*Graph, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, fmt.Errorf("file is empty or missing vertex count")
	}

	nrVertices, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	g := NewGraph(nrVertices)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		from, _ := strconv.Atoi(parts[0])
		to, _ := strconv.Atoi(parts[1])
		g.AddEdge(from, to)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return g, nil
}
