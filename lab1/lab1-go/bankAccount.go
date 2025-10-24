package main

import (
	"math/rand"
	"sync"
	"time"
)

type BankAccount struct {
	ID      int
	Balance int
	Ch      chan Transaction
	mu      sync.Mutex
}

func (account *BankAccount) Start(wg *sync.WaitGroup) {
	go func() {
		for tran := range account.Ch {

			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
			if account.ID == tran.From {
				// fmt.Printf("Account %d: with old balance: %.2f trying to take amount %.2f\n", account.ID, account.Balance, tran.Amount)
				if account.Balance >= tran.Amount {
					tran.otherTxDoneCh <- true
					account.mu.Lock()
					account.Balance -= tran.Amount
					account.mu.Unlock()
				} else {
					// fmt.Printf("Account %d: insufficient funds for %d\n", account.ID, tran.Amount)
					tran.otherTxDoneCh <- false
				}

			} else if account.ID == tran.To {
				// fmt.Printf("Account %d: with old balance: %d trying to put amount %d\n", account.ID, account.Balance, tran.Amount)
				success := <-tran.otherTxDoneCh
				if success {
					account.mu.Lock()
					account.Balance += tran.Amount
					account.mu.Unlock()

				}
			}
			// fmt.Printf("Account %d: balance %d\n", account.ID, account.Balance)
			wg.Done()
		}
	}()
}
