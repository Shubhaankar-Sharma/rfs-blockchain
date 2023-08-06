package blockchain

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

type Block struct {
	Creator    Address        `json:"creator"`
	Height     uint64         `json:"height"`
	Hash       string         `json:"hash"`
	Nonce      uint64         `json:"nonce"`
	PrevHash   string         `json:"prev_hash"`
	Operations []OperationMsg `json:"operations"`
}

func NewBlock(creator Address, height uint64, prevHash string, operations []OperationMsg) Block {
	return Block{
		Creator:    creator,
		Height:     height,
		PrevHash:   prevHash,
		Operations: operations,
	}
}

func (b *Block) ValidateBlock(currHeight uint64, lastHash string) bool {
	// Validate Proof
	if b.Height != (currHeight + 1) {
		return false
	}

	if b.PrevHash != lastHash {
		return false
	}

	if b.Hash != b.GenerateHash() {
		return false
	}

	// Validate Operations

	proof := NewProof(*b, Difficulty)
	return proof.ValidateProof()

}

func (b *Block) IsGenesisBlock() bool {
	return b.Hash == GenesisBlockHash &&
		b.PrevHash == "0" &&
		b.Height == 0 &&
		b.Nonce == 0 &&
		b.Creator == "0" &&
		len(b.Operations) == 0
}

func (b *Block) GenerateHash() string {
	hash := md5.New()
	hash.Write([]byte(fmt.Sprintf("%v%v", b.PrevHash, b.Nonce)))
	return hex.EncodeToString(hash.Sum(nil))
}

func (b *Block) SetNonce(nonce uint64) {
	b.Nonce = nonce
}

func (b *Block) SetHash(hash string) {
	b.Hash = hash
}

func MineBlock(block Block, difficulty uint64) (Block, error) {
	proof := NewProof(block, difficulty)
	block, err := proof.MineBlock()
	if err != nil {
		return Block{}, err
	}
	return block, nil
}

// func (b *Block)
// generate genesis block
func GenerateGenesisBlock() Block {
	return Block{
		Height:     0,
		Creator:    "0",
		Hash:       GenesisBlockHash,
		Nonce:      0,
		PrevHash:   "0",
		Operations: []OperationMsg{},
	}

}
