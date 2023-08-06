package main

import (
	"encoding/binary"
	"fmt"

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
}
