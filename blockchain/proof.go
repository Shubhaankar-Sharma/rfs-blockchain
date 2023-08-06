package blockchain

import (
	"errors"
	"math"
	"math/big"
)

type Proof struct {
	block  Block    `json:"block"`
	target *big.Int `json:"difficulty"`
}

func NewProof(block Block, difficulty uint64) Proof {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficulty))
	return Proof{
		block:  block,
		target: target,
	}
}

func (p *Proof) ValidateProof() bool {
	if p.block.GenerateHash() != p.block.Hash {
		return false
	}

	hashInt := new(big.Int)
	hashInt.SetString(p.block.Hash, 16)
	// check if hashInt < difficulty
	return hashInt.Cmp(p.target) == -1
}

func (p *Proof) MineBlock() (Block, error) {
	nonce := uint64(0)

	for nonce < math.MaxUint64 {

		p.block.SetNonce(nonce)
		p.block.SetHash(p.block.GenerateHash())
		// fmt.Println("mining block", p.block.Height, "with nonce", nonce, "and hash", p.block.Hash)
		if p.ValidateProof() {
			return p.block, nil
		}

		nonce++
	}

	return Block{}, errors.New("block failed to mine")
}
