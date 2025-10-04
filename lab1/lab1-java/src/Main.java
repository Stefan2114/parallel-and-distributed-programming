void main() {


    int bankAccountsNr = 20;
    int transactionsNr = 1000;
    double maxTransferAmount = 300;
    double maxAccountStartBalance = 2000;

    Random rand = new Random();

    List<BankAccount> bankAccounts = new ArrayList<>(bankAccountsNr);
    for (int i = 0; i < bankAccountsNr; i++) {
        bankAccounts.add(new BankAccount(i + 1, BigDecimal.valueOf(rand.nextDouble() * maxAccountStartBalance)));
    }

    BigDecimal allAccountsBalance = BigDecimal.valueOf(0);
    IO.println("First balances:");
    for (BankAccount bankAccount : bankAccounts) {
        IO.println(bankAccount.toString());
        allAccountsBalance = allAccountsBalance.add(bankAccount.getBalance());
    }
    IO.println("All accounts balance: " + allAccountsBalance);

    List<Thread> threads = new ArrayList<>(transactionsNr);
    for (int i = 0; i < transactionsNr; i++) {
        BankAccount from = bankAccounts.get(rand.nextInt(bankAccountsNr));
        BankAccount to;
        do {
            to = bankAccounts.get(rand.nextInt(bankAccountsNr));
        } while (to.equals(from));
        BigDecimal amount = BigDecimal.valueOf(rand.nextDouble() * maxTransferAmount);
        Transaction transaction = new Transaction(from, to, amount);
        Thread thread = Thread.ofVirtual().unstarted(transaction);
        threads.add(thread);
    }

    IO.println("Starting transaction processing...");
    long startTime = System.nanoTime();

    for (int i = 0; i < transactionsNr; i++) {
        Thread thread = threads.get(i);
        thread.start();
    }

    for (int i = 0; i < transactionsNr; i++) {
        Thread thread = threads.get(i);
        try {
            thread.join();
        } catch (InterruptedException e) {
            IO.println(e.getMessage());
        }
    }

    long endTime = System.nanoTime(); // <-- End timer
    allAccountsBalance = BigDecimal.valueOf(0);
    IO.println("Final balances:");
    for (BankAccount bankAccount : bankAccounts) {
        IO.println(bankAccount.toString());
        allAccountsBalance = allAccountsBalance.add(bankAccount.getBalance());
    }
    IO.println("All accounts balance: " + allAccountsBalance);

    long durationNs = endTime - startTime;
    double durationMs = durationNs / 1_000_000.0;
    IO.println("Transaction processing took: " + durationMs + " ms");
}
// All accounts balance: 20269.94173576984204
// All accounts balance: 20269.94173576984204000