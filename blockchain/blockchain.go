package blockchain

import "sync"

type BlockChain struct {
	CurrentHeight    uint64         `json:"current_height"`
	rwMutex          sync.RWMutex   `json:"rw_mutex"`
	Blocks           []Block        `json:"blocks"`
	FileStore        RFSStore       `json:"file_store"`
	operationMemPool []OperationMsg `json:"operation_mem_pool"`
}

// run blockchain
func (bc *BlockChain) Run() {
	if len(bc.Blocks) == 0 {
		genesis := GenerateGenesisBlock()
		bc.AddBlock(genesis)
	}

	// run miner
	// the miner will constantly be adding blocks to the blockchain
	// if it gets a signal for an operation, it will start timer, and will not generate new jobs
	// when timer ends, it will generate a new block with the operation

	// run operation handler
	// the operation handler will constantly be listening for new operations
	// and will add to the mempool

	// block listener will constantly be listening for new blocks
	// and will add to the if the block is valid blockchain

}

func (bc *BlockChain) AddBlock(block Block) {
	bc.rwMutex.Lock()
	defer bc.rwMutex.Unlock()
	if len(bc.Blocks) == 0 {
		bc.Blocks = append(bc.Blocks, block)
		bc.CurrentHeight++
		return
	}

	if block.ValidateBlock(bc.CurrentHeight, bc.Blocks[bc.CurrentHeight-1].Hash) {
		bc.Blocks = append(bc.Blocks, block)
		bc.CurrentHeight++
	}
}

// mine block

// validate block

// validate operation

// validate operations

// validate file
