package main

type Transaction struct {
	From          int
	To            int
	Amount        int
	otherTxDoneCh chan bool
}
