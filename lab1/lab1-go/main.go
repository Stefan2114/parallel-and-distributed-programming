package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func checkBalance(accounts []*BankAccount, initialTotal int, done chan struct{}) {
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				currentTotal := 0
				for _, acc := range accounts {
					acc.mu.Lock()
					currentTotal += acc.Balance
					acc.mu.Unlock()
				}
				if currentTotal-initialTotal != 0 {
					fmt.Printf("Total mismatch! Expected %d, found %.d\n", initialTotal, currentTotal)
				} else {
					fmt.Printf("Total consistent: %.d\n", currentTotal)
				}
			case <-done:
				return
			}
		}
	}()

}

func main() {
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup
	accNr := 20
	tranNr := 1000
	maxTranAmount := 500
	maxAccStartBalance := 1000
	accounts := make([]*BankAccount, accNr)

	for i := 0; i < accNr; i++ {
		accounts[i] = &BankAccount{
			ID:      i + 1,
			Balance: rand.Intn(maxAccStartBalance),
			Ch:      make(chan Transaction, tranNr),
		}
		accounts[i].Start(&wg)
	}

	fmt.Println("\nStart balances:")
	total := 0
	for _, acc := range accounts {
		fmt.Printf("Account %d: %d\n", acc.ID, acc.Balance)
		total += acc.Balance
	}
	fmt.Printf("Total balance: %d\n\n", total)

	start := time.Now() // record start time
	for i := 0; i < tranNr; i++ {
		fromId := rand.Intn(accNr)
		toId := rand.Intn(accNr)
		for fromId == toId {
			toId = rand.Intn(accNr)
		}

		fromAccount := accounts[fromId]
		toAccount := accounts[toId]

		tran := Transaction{
			From:          fromAccount.ID,
			To:            toAccount.ID,
			Amount:        rand.Intn(maxTranAmount),
			otherTxDoneCh: make(chan bool),
		}

		// fmt.Printf("Starting new transaction between account %d: and account: %d with amount %d\n", fromAccount.ID, toAccount.ID, tran.Amount)
		wg.Add(2)
		fromAccount.Ch <- tran
		toAccount.Ch <- tran
	}

	done := make(chan struct{})
	checkBalance(accounts, total, done)
	wg.Wait()
	done <- struct{}{}
	end := time.Since(start)

	for _, acc := range accounts {
		close(acc.Ch)
	}

	fmt.Println("\nFinal balances:")
	total = 0
	for _, acc := range accounts {
		fmt.Printf("Account %d: %d\n", acc.ID, acc.Balance)
		total += acc.Balance
	}
	fmt.Printf("Total balance: %d\n", total)
	fmt.Printf("Transaction processing took: %d ms", end.Milliseconds())

}
