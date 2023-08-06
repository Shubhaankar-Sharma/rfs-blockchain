package blockchain

import (
	"errors"
	"math"
	"math/big"
)

type Proof struct {
	block      Block  `json:"block"`
	difficulty uint64 `json:"difficulty"`
}

func NewProof(block Block, difficulty uint64) Proof {
	return Proof{
		block:      block,
		difficulty: difficulty,
	}
}

func (p *Proof) ValidateProof() bool {
	if p.block.GenerateHash() != p.block.Hash {
		return false
	}

	hashInt := new(big.Int)
	hashInt.SetString(p.block.Hash, 16)

	// check if hashInt < difficulty
	return hashInt.Cmp(big.NewInt(int64(p.difficulty))) == -1
}

func (p *Proof) MineBlock() (Block, error) {
	nonce := uint64(0)

	for nonce < math.MaxUint64 {
		p.block.SetNonce(nonce)
		p.block.SetHash(p.block.GenerateHash())

		if p.ValidateProof() {
			return p.block, nil
		}

		nonce++
	}

	return Block{}, errors.New("block failed to mine")
}
