package blockchain

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Shubhaankar-Sharma/rfs-blockchain/crypto"
)

type MINING_STATE int

const (
	IDLE MINING_STATE = iota
	NO_OPERATION_MINING
	OPERATION_MINING
	WAITING_FOR_OPERATION
)

func (s MINING_STATE) String() string {
	return [...]string{"IDLE", "NO_OPERATION_MINING", "OPERATION_MINING", "WAITING_FOR_OPERATION"}[s]
}

type BlockChain struct {
	Wallet           *crypto.Account `json:"wallet"`
	CurrentHeight    uint64          `json:"current_height"`
	rwMutex          sync.RWMutex    `json:"rw_mutex"`
	blocks           []Block         `json:"blocks"`
	fileStore        RFSStore        `json:"file_store"`
	ledger           LedgerStore     `json:"ledger"`
	operationMemPool []OperationMsg  `json:"operation_mem_pool"`
	miningState      struct {
		state         MINING_STATE
		timeStartChan chan struct{}
		timeOutChan   chan struct{}
		abandonMining chan struct{}
	}
}

func NewBlockchain() *BlockChain {
	wallet := crypto.NewAccount()
	return &BlockChain{
		CurrentHeight:    0,
		blocks:           []Block{},
		fileStore:        RFSStore{Files: map[string]File{}},
		ledger:           LedgerStore{Ledger: map[Address]*AccountStorage{}},
		operationMemPool: []OperationMsg{},
		Wallet:           wallet,
		miningState: struct {
			state         MINING_STATE
			timeStartChan chan struct{}
			timeOutChan   chan struct{}
			abandonMining chan struct{}
		}{
			state:         IDLE,
			timeStartChan: make(chan struct{}),
			timeOutChan:   make(chan struct{}),
			abandonMining: make(chan struct{}),
		},
	}
}

// run blockchain
func (bc *BlockChain) Run() {
	if len(bc.blocks) == 0 {
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
	bc.runMiner()
}

func (bc *BlockChain) runMiner() {
	// how do we do this...
	// lets have a

	var noOpCtx, opCtx context.Context
	var noOpCancel, opCancel context.CancelFunc
	// TODO: replace context.Backround with service level context
	noOpCtx, noOpCancel = context.WithCancel(context.Background())
	opCtx, opCancel = context.WithCancel(context.Background())

	noOpSig := make(chan struct{}, 1)
	// start mining
	noOpSig <- struct{}{}
	go func() {
		for {
			select {
			case <-bc.miningState.abandonMining:
				fmt.Println("abandoning mining")
				noOpCancel()
				opCancel()
				return
				// case <-ctx.Done()
			}
		}
	}()

	for {
		fmt.Printf("Blockchain Current Height: %v \n", bc.CurrentHeight)
		select {
		case <-bc.miningState.timeStartChan:
			opCtx, opCancel = context.WithCancel(context.Background())
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func(ctx context.Context) {
				bc.setState(WAITING_FOR_OPERATION)
				defer bc.setState(IDLE)
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					// wait till timeout to mine
					case <-bc.miningState.timeOutChan:
						bc.setState(OPERATION_MINING)
						opBlock := bc.InitOpBlock()
						opBlock, err := MineBlock(opBlock, Difficulty, ctx.Done())
						if err == nil {
							bc.AddBlock(opBlock)
						}
						return
						// mine with mempool
					}

				}
			}(opCtx)
			wg.Wait()
		case <-noOpSig:
			noOpCtx, noOpCancel = context.WithCancel(context.Background())
			func() {
				// mine no-op block
				noOpBlock := bc.InitNoOpBlock()
				defer func() { noOpSig <- struct{}{} }()
				defer bc.setState(IDLE)
				bc.setState(NO_OPERATION_MINING)
				noOpBlock, err := MineBlock(noOpBlock, Difficulty, noOpCtx.Done())
				if err != nil {
					return
				}
				bc.AddBlock(noOpBlock)
			}()
		}
	}
}

func (bc *BlockChain) InitNoOpBlock() Block {
	bc.rwMutex.RLock()
	defer bc.rwMutex.RUnlock()
	block := Block{
		Height:     bc.CurrentHeight + 1,
		Creator:    Address(bc.Wallet.Address),
		PrevHash:   bc.blocks[bc.CurrentHeight-1].Hash,
		Operations: []OperationMsg{},
	}
	return block
}

func (bc *BlockChain) InitOpBlock() Block {
	bc.rwMutex.Lock()
	defer bc.rwMutex.Unlock()
	operations := make([]OperationMsg, len(bc.operationMemPool))
	copy(operations, bc.operationMemPool)
	bc.operationMemPool = []OperationMsg{}

	block := Block{
		Height:     bc.CurrentHeight + 1,
		Creator:    Address(bc.Wallet.Address),
		PrevHash:   bc.blocks[bc.CurrentHeight-1].Hash,
		Operations: operations,
	}
	return block
}

func (bc *BlockChain) AddBlock(block Block) {
	bc.rwMutex.Lock()
	if len(bc.blocks) == 0 {
		bc.blocks = append(bc.blocks, block)
		bc.CurrentHeight++
		bc.rwMutex.Unlock()
		return
	}

	if block.ValidateBlock(bc.CurrentHeight, bc.blocks[bc.CurrentHeight-1].Hash) {
		bc.blocks = append(bc.blocks, block)
		bc.CurrentHeight++
		bc.rwMutex.Unlock()
		bc.processBlockLedgerAndOperations(block)
	}
}

func (bc *BlockChain) processBlockLedgerAndOperations(block Block) {
	bc.rwMutex.Lock()
	defer bc.rwMutex.Unlock()
	// process ledger
	// give reward
	reward := uint64(MinedCoinsPerNoOpBlock)
	if len(block.Operations) > 0 {
		reward = uint64(MinedCoinsPerOpBlock)
	}

	if bc.ledger.Ledger[block.Creator] == nil {
		bc.ledger.Ledger[block.Creator] = &AccountStorage{
			Balance: reward,
			Nonce:   0,
		}
	} else {
		bc.ledger.Ledger[block.Creator].Balance += reward
	}
	fmt.Printf("processed Block: %v, rewarded: %v, balance: %v \n", block, block.Creator, bc.ledger.Ledger[block.Creator].Balance)
	// process operations
}

func (bc *BlockChain) AddOperation(op OperationMsg) {
	// validate operation
	if ValidateOperation(op) {
		bc.rwMutex.Lock()
		bc.operationMemPool = append(bc.operationMemPool, op)
		bc.rwMutex.Unlock()

		if bc.readState() == WAITING_FOR_OPERATION {
			return
		}
		bc.startOpMiningTimer()
	}
}

// TODO: USE ATOMIC
func (bc *BlockChain) setState(state MINING_STATE) {
	bc.rwMutex.Lock()
	defer bc.rwMutex.Unlock()
	bc.miningState.state = state
	fmt.Println("state changed to: ", state.String())
}

func (bc *BlockChain) readState() MINING_STATE {
	bc.rwMutex.RLock()
	defer bc.rwMutex.RUnlock()
	return bc.miningState.state
}

func (bc *BlockChain) startOpMiningTimer() {
	go func() {
		bc.miningState.timeStartChan <- struct{}{}
		<-time.After(GenOpBlockTimeout)
		bc.miningState.timeOutChan <- struct{}{}
	}()
}
