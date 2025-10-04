import java.math.BigDecimal;
import java.util.Random;
import java.util.concurrent.locks.ReentrantLock;

public class Transaction implements Runnable {

    private final BankAccount fromAccount;
    private final BankAccount toAccount;
    private final BigDecimal amount;
    private final Random rand;
    private static final int CHANCE_OF_FAILING = 2;

    public Transaction(BankAccount fromAccount, BankAccount toAccount, BigDecimal amount) {
        this.fromAccount = fromAccount;
        this.toAccount = toAccount;
        this.amount = amount;
        this.rand = new Random();
    }

    @Override
    public void run() {
        try {
            Thread.sleep((int) (rand.nextDouble() * 1000));
        } catch (InterruptedException ignored) {
        }
        int fromAccountID = this.fromAccount.getId();
        int toAccountID = this.toAccount.getId();
        IO.println("Account with id: " + fromAccountID + " current balance: " + this.fromAccount.getBalance());
        IO.println("Account with id: " + toAccountID + " current balance: " + this.toAccount.getBalance());
        IO.println("Trying to transfer the amount: " + this.amount);
        ReentrantLock fromAccountLock = this.fromAccount.getLock();
        ReentrantLock toAccountLock = this.toAccount.getLock();
        boolean condition = BankAccount.LOCKS.indexOf(fromAccountLock) < BankAccount.LOCKS.indexOf(toAccountLock);
        ReentrantLock first = condition ? fromAccountLock : toAccountLock;
        ReentrantLock second = condition ? this.toAccount.getLock() : this.fromAccount.getLock();
        first.lock();
        second.lock();
        if (this.fromAccount.getBalance().compareTo(this.amount) < 0) {
            IO.println("Account with id: " + fromAccountID
                    + " doesn't have enough money for the transaction. Account amount: "
                    + this.fromAccount.getBalance() + ", amount trying to take: " + this.amount);
            first.unlock();
            second.unlock();
            return;
        }
        try {
            makeTransaction(this.fromAccount, this.toAccount, amount);
        } catch (Exception e) {
            IO.println(e.getMessage());
        } finally {
            IO.println("Account with id: " + fromAccountID + " current balance: " + this.fromAccount.getBalance());
            IO.println("Account with id: " + toAccountID + " current balance: " + this.toAccount.getBalance());
            first.unlock();
            second.unlock();
        }
    }

    private void makeTransaction(BankAccount from, BankAccount to, BigDecimal amount) throws Exception {
        int randomNr = rand.nextInt(100);
        if (randomNr < CHANCE_OF_FAILING) {
            throw new Exception("Something went wrong in updating amount in Account with id: " + from.getId() + "-------------------------------------------");
        }
        BigDecimal fromNewAmount = from.getBalance().subtract(amount);
        randomNr = rand.nextInt(100);
        if (randomNr < CHANCE_OF_FAILING) {
            throw new Exception("Something went wrong in updating amount in Account with id: " + to.getId() + "-------------------------------------------");
        }
        BigDecimal toNewAmount = to.getBalance().add(amount);
        from.setBalance(fromNewAmount);
        to.setBalance(toNewAmount);
    }
}
