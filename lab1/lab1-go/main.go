package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Transaction struct {
	From          int
	To            int
	Amount        float64
	otherTxDoneCh chan bool
}

type BankAccount struct {
	ID      int
	Balance float64
	Ch      chan Transaction
}

func (account *BankAccount) Start(wg *sync.WaitGroup) {
	go func() {
		for tran := range account.Ch {

			if account.ID == tran.From {
				fmt.Printf("Account %d: with old balance: %.2f trying to take amount %.2f\n", account.ID, account.Balance, tran.Amount)
				if account.Balance >= tran.Amount {
					account.Balance -= tran.Amount
					tran.otherTxDoneCh <- true
				} else {
					fmt.Printf("Account %d: insufficient funds for %.2f\n", account.ID, tran.Amount)
					tran.otherTxDoneCh <- false
				}

			} else if account.ID == tran.To {
				fmt.Printf("Account %d: with old balance: %.2f trying to put amount %.2f\n", account.ID, account.Balance, tran.Amount)
				success := <-tran.otherTxDoneCh
				if success {
					account.Balance += tran.Amount
				}
			}
			fmt.Printf("Account %d: balance %.2f\n", account.ID, account.Balance)

			wg.Done()
		}
	}()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup
	accNr := 20
	tranNr := 1000
	maxTranAmount := 500.0
	maxAccStartBalance := 1000.0
	accounts := make([]*BankAccount, accNr)

	for i := 0; i < accNr; i++ {
		accounts[i] = &BankAccount{
			ID:      i + 1,
			Balance: rand.Float64() * maxAccStartBalance,
			Ch:      make(chan Transaction, tranNr),
		}
		accounts[i].Start(&wg)
	}

	fmt.Println("\nStart balances:")
	total := 0.0
	for _, acc := range accounts {
		fmt.Printf("Account %d: %.2f\n", acc.ID, acc.Balance)
		total += acc.Balance
	}
	fmt.Printf("Total balance: %.2f\n\n", total)

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
			Amount:        rand.Float64() * maxTranAmount,
			otherTxDoneCh: make(chan bool),
		}

		fmt.Printf("Starting new transaction bettwen account %d: and account: %d with amount %.2f\n", fromAccount.ID, toAccount.ID, tran.Amount)

		wg.Add(2)
		fromAccount.Ch <- tran
		toAccount.Ch <- tran
	}

	wg.Wait()
	end := time.Since(start) // calculate duration

	for _, acc := range accounts {
		close(acc.Ch)
	}

	fmt.Println("\nFinal balances:")
	total = 0.0
	for _, acc := range accounts {
		fmt.Printf("Account %d: %.2f\n", acc.ID, acc.Balance)
		total += acc.Balance
	}
	fmt.Printf("Total balance: %.2f\n", total)
	fmt.Printf("Transaction processing took: %d ms", end.Milliseconds())

}
