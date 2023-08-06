package blockchain

import "sync"

type BlockChain struct {
	CurrentHeight    uint64         `json:"current_height"`
	rwMutex          *sync.RWMutex  `json:"rw_mutex"`
	Blocks           []Block        `json:"blocks"`
	FileStore        RFSStore       `json:"file_store"`
	operationMemPool []OperationMsg `json:"operation_mem_pool"`
}

// run blockchain
func (bc *BlockChain) Run() {
	// TODO
}

// mine block

// validate block

// validate operation

// validate operations

// validate file
