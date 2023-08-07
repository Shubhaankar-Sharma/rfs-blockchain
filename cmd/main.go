package main

import (
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
		bc.AddOperation(blockchain.OperationMsg{
			OpType: blockchain.CREATE_FILE,
		})
		time.Sleep(time.Second * 1)
	}
	wg.Wait()
}
