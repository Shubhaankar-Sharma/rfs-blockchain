package network

import "github.com/Shubhaankar-Sharma/rfs-blockchain/blockchain"

// var _ IBlockchain = (*blockchain.BlockChain)(nil)

type IBlockchain interface {
	CurrentHeight() uint64
	GetBlockPubSubChans() chan<- blockchain.Block
	GetOperationPubSubChans() chan<- blockchain.OperationMsg
	// GetBlockByNumber(uint64) (blockchain.Block, error)
	// GetOperationPool() map[string]blockchain.OperationMsg
	// GetBlocksByIndex(uint64, uint64) ([]blockchain.Block, error)
	GetFullBlockchain() []blockchain.Block
	GetMemPool() map[string]blockchain.OperationMsg
	GetBlock(uint64) (blockchain.Block, error)
}
