import java.math.BigDecimal;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.locks.ReentrantLock;

public class BankAccount {
    private final int id;
    private BigDecimal balance;
    private final ReentrantLock lock;
    public static List<ReentrantLock> LOCKS = new ArrayList<>();

    public BankAccount(int id, BigDecimal balance) {
        this.id = id;
        this.balance = balance;
        this.lock = new ReentrantLock();
        this.addLock(this.lock);
    }

    private synchronized void addLock(ReentrantLock lock) {
        LOCKS.add(lock);
    }

    public int getId() {
        return this.id;
    }

    public BigDecimal getBalance() {
        return balance;
    }

    public void setBalance(BigDecimal balance) {
        this.balance = balance;
    }

    public ReentrantLock getLock() {
        return this.lock;
    }

    @Override
    public boolean equals(Object other) {
        if (other == null) {
            return false;
        }
        if (!(other instanceof BankAccount otherAccount)) {
            return false;
        }
        return this.id == otherAccount.id;
    }

    @Override
    public String toString() {
        return "Bank account with id: " + id + " has balance: " + balance;
    }
}
