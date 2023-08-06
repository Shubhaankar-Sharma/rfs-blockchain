package main

import (
	"encoding/binary"
	"fmt"

	"github.com/Shubhaankar-Sharma/rfs-blockchain/blockchain"
	"github.com/Shubhaankar-Sharma/rfs-blockchain/crypto"
)

func main() {
	account := crypto.NewAccount()
	account.IncrementNonce()
	msg := []byte("hello world")
	hash, sig, _ := account.Sign(msg)
	nonceBytes := hash[:binary.MaxVarintLen64]
	nonceD, _ := binary.Uvarint(nonceBytes)
	fmt.Println(len(sig), nonceD)

	block, err := blockchain.MineBlock(
		blockchain.NewBlock(blockchain.Address(account.Address), 0, blockchain.GenesisBlockHash, nil),
		blockchain.Difficulty,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(block)

	for i := 0; i < 20; i++ {
		block = blockchain.NewBlock(blockchain.Address(account.Address), uint64(i+1), block.Hash, nil)
		block, err = blockchain.MineBlock(block, blockchain.Difficulty)
		if err != nil {
			panic(err)
		}
		fmt.Println(block)
		fmt.Println(block.ValidateBlock(uint64(i), block.PrevHash))
	}
}
