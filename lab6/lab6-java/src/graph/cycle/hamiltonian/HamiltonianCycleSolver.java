package graph.cycle.hamiltonian;

import graph.Graph;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ForkJoinPool;
import java.util.concurrent.RecursiveTask;
import java.util.concurrent.atomic.AtomicBoolean;

public class HamiltonianCycleSolver {

    private final Graph graph;
    private final AtomicBoolean solutionFound = new AtomicBoolean(false);

    public HamiltonianCycleSolver(Graph graph) {
        this.graph = graph;
    }

    public List<Integer> solve(int startVertex, int nrThreads) {
        solutionFound.set(false);
        try (ForkJoinPool pool = new ForkJoinPool(nrThreads)) {
            List<Integer> path = new ArrayList<>(graph.getNrVertices());
            path.add(startVertex);
            boolean[] visited = new boolean[graph.getNrVertices()];
            visited[startVertex] = true;

            return pool.invoke(new SearchTask(path, visited, startVertex, nrThreads));
        }
    }

    private class SearchTask extends RecursiveTask<List<Integer>> {
        private final List<Integer> path;
        private final boolean[] visited;
        private final int currentVertex;
        private final int allocatedThreads;

        public SearchTask(List<Integer> path, boolean[] visited, int currentVertex, int allocatedThreads) {
            this.path = path;
            this.visited = visited;
            this.currentVertex = currentVertex;
            this.allocatedThreads = allocatedThreads;
        }

        @Override
        protected List<Integer> compute() {

            if (solutionFound.get()) return null;
            if (this.path.size() == graph.getNrVertices()) {
                return checkSolution(this.currentVertex);
            }

            List<Integer> validNeighbors = getValidNeighbors();
            if (validNeighbors.isEmpty()) return null;

            if (this.allocatedThreads <= 1) {
                return searchSequential(this.currentVertex, this.path, visited);
            }

            return searchParallel(validNeighbors);
        }

        private List<Integer> searchSequential(int currentVertex, List<Integer> path, boolean[] visited) {

            if (solutionFound.get()) return null;

            if (path.size() == graph.getNrVertices()) {
                return checkSolution(currentVertex);
            }

            for (int neighbor : graph.getNeighbors(currentVertex)) {
                if (!visited[neighbor]) {
                    visited[neighbor] = true;
                    path.add(neighbor);

                    List<Integer> res = searchSequential(neighbor, path, visited);
                    if (res != null) return res;

                    path.removeLast();
                    visited[neighbor] = false;
                }
            }
            return null;
        }

        private List<Integer> searchParallel(List<Integer> validNeighbors){
            List<SearchTask> subTasks = new ArrayList<>();
            int splitThreads = Math.max(1, this.allocatedThreads / validNeighbors.size());

            for (int neighbor : validNeighbors) {
                if (solutionFound.get()) return null;

                List<Integer> newPath = new ArrayList<>(this.path);
                newPath.add(neighbor);
                boolean[] newVisited = visited.clone();
                newVisited[neighbor] = true;

                SearchTask task = new SearchTask(newPath, newVisited, neighbor, splitThreads);
                task.fork();
                subTasks.add(task);
            }

            for (SearchTask task : subTasks) {
                List<Integer> result = task.join();
                if (result != null) return result;
            }

            return null;
        }

        private List<Integer> checkSolution(int vertex){


                if (graph.hasEdge(vertex, this.path.getFirst())) {
                    if (solutionFound.compareAndSet(false, true)) {
                        return new ArrayList<>(this.path);
                    }
                }
                return null;
        }


        private List<Integer> getValidNeighbors() {

            List<Integer> validNeighbors = new ArrayList<>();
            for (int neighbor : graph.getNeighbors(this.currentVertex)) {
                if (!visited[neighbor]) {
                    validNeighbors.add(neighbor);
                }
            }

            return validNeighbors;
        }
    }
}