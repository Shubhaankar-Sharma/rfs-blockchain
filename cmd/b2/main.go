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
	n := network.NewNetwork("localhost:1235", b, "localhost:1234")
	n.RegisterChannels()
	go func() {
		err := n.Run()
		if err != nil {
			panic(err)
		}
	}()
	// go b.Run()
	for ; ; _ = <-time.NewTicker(time.Second * 2).C {
		bestHeight, addy := n.GetBestHeight()
		fmt.Printf("Best Height From Peers: %d %s \n", bestHeight, addy)
	}
	// fmt.Println(n.GetBestHeight())
}
