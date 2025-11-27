import graph.Graph;
import graph.cycle.hamiltonian.HamiltonianCycleSolver;

void main() throws IOException {

    //    int nrVertices = 35;
    //    int nrEdges = 4 *  nrVertices;
    int nrThreads = 100;
    int startVertex = 0;
    Graph g = Graph.readFromFile("graph2.txt");

    HamiltonianCycleSolver solver = new HamiltonianCycleSolver(g);

    long start = System.currentTimeMillis();
    List<Integer> cycle = solver.solve(startVertex, nrThreads);
    long end = System.currentTimeMillis();

    if (cycle != null) {
        System.out.println("Hamiltonian Cycle found!");
        System.out.println(cycle);

    } else {
        System.out.println("No cycle found.");
    }

    System.out.println("Time taken: " + (end - start) + "ms");
}

