package main

import (
	"fmt"

	"github.com/Shubhaankar-Sharma/rfs-blockchain/crypto"
)

func main() {
	fmt.Println(len(crypto.NewAccount().PublicKey))
}
