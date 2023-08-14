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
	n := network.NewNetwork("localhost:1234", b)
	n.RegisterChannels()
	go func() {
		err := n.Run()
		if err != nil {
			panic(err)
		}
	}()
	go b.Run()
	ticker := time.NewTicker(time.Second * 2)
	for ; ; <-ticker.C {
		bestHeight, addy := n.GetBestHeight()
		fmt.Printf("Best Height From Peers: %d %s \n", bestHeight, addy)
	}
	// fmt.Println(n.GetBestHeight())
}
