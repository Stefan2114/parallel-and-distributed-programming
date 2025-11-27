package graph;

import java.io.BufferedReader;
import java.io.FileReader;
import java.io.IOException;
import java.util.*;

public class Graph {

    private final int nrVertices;
    private final Map<Integer, List<Integer>> adjList;

    public Graph(int nrVertices) {
        this.nrVertices = nrVertices;
        this.adjList = new HashMap<>();
        for (int i = 0; i < nrVertices; i++) {
            adjList.put(i, new ArrayList<>());
        }
    }

    public void addEdge(int v1, int v2) {
        adjList.get(v1).add(v2);
    }

    public int getNrVertices() {
        return nrVertices;
    }

    public List<Integer> getNeighbors(int v1) {
        return adjList.get(v1);
    }

    public boolean hasEdge(int v1, int v2) {
        return adjList.get(v1).contains(v2);
    }

    public List<List<Integer>> toList(){
        List<List<Integer>> list = new ArrayList<>(this.nrVertices);
        for (int i = 0; i < nrVertices; i++) {
            list.add(this.adjList.get(i));
        }
        return list;
    }


    public static Graph readFromFile(String filename) throws IOException {
        try (BufferedReader br = new BufferedReader(new FileReader(filename))) {

            String line = br.readLine();
            if (line == null) {
                throw new IOException("File is empty");
            }
            int n = Integer.parseInt(line.trim());
            Graph graph = new Graph(n);

            while ((line = br.readLine()) != null) {
                line = line.trim();
                if (line.isEmpty()) continue;

                String[] parts = line.split("\\s+");

                if (parts.length >= 2) {
                    int from = Integer.parseInt(parts[0]);
                    int to = Integer.parseInt(parts[1]);

                    graph.addEdge(from, to);
                }
            }
            return graph;
        }
    }
}

