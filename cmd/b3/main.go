package main

import (
	"fmt"
	"time"

	"github.com/Shubhaankar-Sharma/rfs-blockchain/blockchain"
	"github.com/Shubhaankar-Sharma/rfs-blockchain/network"
)

func main() {
	// ...
	// create a new blockchain
	b := blockchain.NewBlockchain()
	n := network.NewNetwork("localhost:1236", b, "localhost:1235")
	n.RegisterChannels()
	go func() {
		err := n.Run()
		if err != nil {
			panic(err)
		}
	}()
	for ; ; <-time.NewTicker(time.Second * 1).C {
		bestHeight, addy := n.GetBestHeight()
		fmt.Printf("-------------Best Height From Peers-----------------: %d %s \n", bestHeight, addy)
	}
	// fmt.Println(n.GetBestHeight())
}
