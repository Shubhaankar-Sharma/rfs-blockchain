package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/Shubhaankar-Sharma/rfs-blockchain/blockchain"
)

func main() {
	bc := blockchain.NewBlockchain()
	bc.Run()
	time.Sleep(time.Second * 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	// add 12 operations to mempool
	optypes := []blockchain.OperationType{
		blockchain.CREATE_FILE,
		blockchain.APPEND_RECORD,
		blockchain.DELETE_FILE,
		blockchain.CREATE_FILE,
		blockchain.APPEND_RECORD,
		blockchain.DELETE_FILE,
		blockchain.CREATE_FILE,
		blockchain.APPEND_RECORD,
		blockchain.DELETE_FILE,
		blockchain.CREATE_FILE,
		blockchain.APPEND_RECORD,
		blockchain.DELETE_FILE,
	}

	for i := 0; i < 12; i++ {
		op, err := bc.SignTransaction(blockchain.OperationMsg{
			OpType: optypes[i],
		})
		op.GenerateHash()
		if err != nil {
			panic(err)
		}
	tryAgain:
		err = bc.AddOperation(op)

		if err != nil {
			fmt.Println("waiting for funds, ERROR: ", err.Error(), "id: ", i)
			time.Sleep(time.Second * 1)
			goto tryAgain
		}

		fmt.Println("added operation to mempool: ", i)
		time.Sleep(time.Second * 1)
	}
	// fmt.Println("done")
	// time.Sleep(time.Second * 10)
	// bc.Stop()
	wg.Wait()

}
