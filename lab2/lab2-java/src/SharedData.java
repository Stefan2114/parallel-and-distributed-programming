import java.util.LinkedList;
import java.util.Queue;
import java.util.concurrent.locks.Condition;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantLock;

public class SharedData {
    private final Lock lock = new ReentrantLock();
    private final Condition hasData = lock.newCondition();
    private final Queue<Integer> queue = new LinkedList<>();
    private boolean done = false;

    // Producer
    public void produce(int value) {
        lock.lock();
        try {
            queue.add(value);
            hasData.signal();
        } finally {
            lock.unlock();
        }
    }

    // Consumer
    public Integer consume() {
        lock.lock();
        try {
            while (queue.isEmpty() && !done) {
                hasData.await();
            }
            return queue.poll();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return null;
        } finally {
            lock.unlock();
        }
    }

    public void setDone() {
        lock.lock();
        try {
            done = true;
            hasData.signalAll();
        } finally {
            lock.unlock();
        }
    }
}
