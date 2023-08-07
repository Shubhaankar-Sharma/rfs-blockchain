package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/Shubhaankar-Sharma/rfs-blockchain/blockchain"
)

func main() {
	bc := blockchain.NewBlockchain()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		bc.Run()
	}()
	time.Sleep(time.Second * 2)
	// add 12 operations to mempool
	for i := 0; i < 12; i++ {
		op, err := bc.SignTransaction(blockchain.OperationMsg{
			OpType: blockchain.CREATE_FILE,
		})
		if err != nil {
			panic(err)
		}
	tryAgain:
		err = bc.AddOperation(op)
		if err != nil && err == blockchain.ErrInsufficientFunds {
			fmt.Println("waiting for funds")
			time.Sleep(time.Second * 5)
			goto tryAgain
		}

		time.Sleep(time.Second * 1)
	}
	fmt.Println("done")
}
