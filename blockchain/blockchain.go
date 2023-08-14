package blockchain

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Shubhaankar-Sharma/rfs-blockchain/crypto"
)

type MINING_STATE int32

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
	ctx     context.Context
	cancel  context.CancelFunc
	Running *int32

	Wallet *crypto.Account

	rwMutex sync.RWMutex

	blocks []Block

	fileStore        *RFSStore
	ledger           *LedgerStore
	operationMemPool *OperationMemPool

	network Network

	miningState struct {
		state                     *int32
		timeStartChan             chan struct{}
		timeOutChan               chan struct{}
		abandonMining             chan struct{}
		startOpMiningTimerRunning *atomic.Int32
		miningPaused              *int32
	}

	// chans
	blockListener     chan Block
	operationListener chan OperationMsg

	blockPublisher     chan Block
	operationPublisher chan OperationMsg
}

func NewBlockchain() *BlockChain {
	ctx, cancel := context.WithCancel(context.Background())
	wallet := crypto.NewAccount()
	startOpMiningTimerRunning := atomic.Int32{}
	startOpMiningTimerRunning.Store(0)
	miningState := int32(IDLE)
	running := int32(0)
	miningPaused := int32(0)

	ledger := NewLedgerStore()

	return &BlockChain{
		ctx:     ctx,
		cancel:  cancel,
		Running: &running,

		blocks:           []Block{},
		fileStore:        &RFSStore{Files: map[string]File{}},
		ledger:           ledger,
		operationMemPool: NewOperationMemPool(ledger.copy()),
		Wallet:           wallet,
		miningState: struct {
			state                     *int32
			timeStartChan             chan struct{}
			timeOutChan               chan struct{}
			abandonMining             chan struct{}
			startOpMiningTimerRunning *atomic.Int32
			miningPaused              *int32
		}{
			state:                     &miningState,
			timeStartChan:             make(chan struct{}, 256),
			timeOutChan:               make(chan struct{}, 256),
			abandonMining:             make(chan struct{}, 256),
			startOpMiningTimerRunning: &startOpMiningTimerRunning,
			miningPaused:              &miningPaused,
		},

		blockListener:      make(chan Block, 256),
		operationListener:  make(chan OperationMsg, 256),
		blockPublisher:     make(chan Block, 256),
		operationPublisher: make(chan OperationMsg, 256),
	}
}

func (bc *BlockChain) SetNetwork(network Network) {
	bc.network = network
}

// run blockchain
func (bc *BlockChain) Run() {
	if bc.IsRunning() {
		return
	}

	// TODO: check if you have any peers, if you do try to get the blockchain from them, before running
	// the miner etc
	// how to sync?
	// check with all peers you have
	// get the peer with the best height
	// then get blocks by height from them...
	// get block 1, 2, 3, 4, 5, 6...
	// for the poc we will get full blockchain at once
	// later on we will use pagination to get 50 blocks at once while syncing...
	// after sync is complete, we will start the miner
	// and open our listeners for txns and blocks
	// if we get a new block and its valid, we will add it to our blockchain and abandon our current blockmining
	// if we get a new txn, we will add it to our mempool and start mining a new block

	// TODO: if peers are not empty
	bc.syncBlockChain()
	// TODO: if len nodes to connect to is 0 we generate genesis block
	if len(bc.blocks) == 0 {
		genesis := GenerateGenesisBlock()
		bc.addBlock(genesis)
	}

	go bc.minerWorker()
	go bc.blockPubSubWorker()
	go bc.operationPubSubWorker()

	atomic.StoreInt32(bc.Running, 1)
}

func (bc *BlockChain) IsRunning() bool {
	return atomic.LoadInt32(bc.Running) == 1
}

func (bc *BlockChain) Stop() {
	bc.cancel()
}

func (bc *BlockChain) InitNoOpBlock() Block {
	bc.rwMutex.RLock()
	defer bc.rwMutex.RUnlock()
	block := Block{
		Height:     uint64(len(bc.blocks)) + 1,
		Creator:    Address(bc.Wallet.Address),
		PrevHash:   bc.blocks[(len(bc.blocks))-1].Hash,
		Operations: map[string]OperationMsg{},
	}
	return block
}

func (bc *BlockChain) InitOpBlock() Block {
	operations := bc.operationMemPool.EmptyMemPool()

	height := bc.CurrentHeight()
	bc.rwMutex.RLock()
	prevHash := bc.blocks[height-1].Hash
	bc.rwMutex.RUnlock()

	block := Block{
		Height:     height + 1,
		Creator:    Address(bc.Wallet.Address),
		PrevHash:   prevHash,
		Operations: operations,
	}

	return block
}

// blockMined is called when a block is mined
// it adds the block, and publishes it to the network
func (bc *BlockChain) blockMined(block Block) {
	bc.addBlock(block)
	// publish block to network
}

func (bc *BlockChain) AddLatestBlock(block Block) {
	// check if block already exists in our stack
	// if it does we skip
	currentHeight := bc.CurrentHeight()

	if currentHeight >= block.Height {
		return
	}

	if currentHeight+1 != block.Height {
		// we are out of sync
		// lets fetch all the blocks before this block

		// step 1: pause mining
		bc.abandonMining()
		bc.pauseMining()

		// step 2: get all the blocks before this block
		bc.syncBlockChain()
		// , gets blocks one by one from the best height peer
		// also gets the operation mempool and syncs it...
		// is a blocking function
		// resume mining after sync is complete
		bc.resumeMining()
		return
	}

	// log error
	// if err == nil we have added the block to our blockchain
	// after that we stop our current mining operations and start mining the new block
	if bc.readState() == OPERATION_MINING || bc.readState() == NO_OPERATION_MINING {
		bc.abandonMining()
	}
	// check if the operations in this block are still left in the mempool, if so remove them

	// publish block to network
}

func (bc *BlockChain) GetBlockPubSubChans() (chan<- Block, <-chan Block) {
	// others will send to our listener, and will listen to our publisher
	return bc.blockListener, bc.blockPublisher
}

func (bc *BlockChain) GetOperationPubSubChans() (chan<- OperationMsg, <-chan OperationMsg) {
	// others will send to our listener, and will listen to our publisher
	return bc.operationListener, bc.operationPublisher
}

// addBlock adds and proceses the block to the blockchain memory array
func (bc *BlockChain) addBlock(block Block) error {
	if bc.CurrentHeight() == 0 {
		bc.rwMutex.Lock()
		bc.blocks = append(bc.blocks, block)
		bc.rwMutex.Unlock()
		return nil
	}

	if !block.ValidateBlock(bc.CurrentHeight(), bc.blocks[bc.CurrentHeight()-1].Hash) {
		return ErrInvalidBlock
	}

	ok, err := bc.validateAndApplyOperations(block)
	if err != nil || !ok {
		return ErrInvalidBlock
	}

	bc.rwMutex.Lock()
	bc.blocks = append(bc.blocks, block)
	bc.rwMutex.Unlock()

	bc.addBlockCleanup()

	return nil
}

func (bc *BlockChain) validateAndApplyOperations(block Block) (bool, error) {
	l := bc.ledger.copy()

	// validate transactions
	for _, op := range block.Operations {
		if v, _ := ValidateOperation(op, op.OpFrom, l.GetAccount(op.OpFrom)); !v {
			return false, ErrInvalidOperation
		}
		err := l.ApplyOperation(op)
		if err != nil {
			return false, err
		}
	}
	// give reward
	reward := uint64(MinedCoinsPerNoOpBlock)
	if len(block.Operations) > 0 {
		reward = uint64(MinedCoinsPerOpBlock)
	}

	l.GetAccount(block.Creator).AddBalance(reward)

	fmt.Printf("processed Block: %v, rewarded: %v, balance: %v \n", block, block.Creator, l.GetAccount(block.Creator).GetBalance())

	// replace the ledger
	bc.rwMutex.Lock()
	bc.ledger = l
	bc.operationMemPool.SetLedger(l.copy())
	// replace rfs
	bc.rwMutex.Unlock()

	return true, nil
}

// WARNING! THIS IS EXTERNAL ONLY FUNCTION, IT IS NOT AN INTERNAL FUNCTION
func (bc *BlockChain) AddOperation(op OperationMsg) error {
	if bc.operationMemPool.Exists(op) {
		return ErrAlreadyInMempool
	}

	// publish operation to network

	// validate operation
	var err error

	// try to add to mempool
	err = bc.operationMemPool.AddOperation(op)

	if err != nil {
		return err
	}

	if bc.readState() == WAITING_FOR_OPERATION {
		return nil
	}

	bc.startOpMiningTimer()

	return nil
}

func (bc *BlockChain) syncBlockChain() {
	// gets best height node
	// gets all blocks till we're in sync
	// gets all transactions from the pool of the best node
}

// for clients that connect to this node, we sign transactions via this node's account
func (bc *BlockChain) SignTransaction(op OperationMsg) (OperationMsg, error) {
	op.OpFrom = Address(bc.Wallet.Address)
	hash, sig, err := bc.Wallet.Sign(op.Op)
	if err != nil {
		return OperationMsg{}, err
	}
	op.Signature = struct {
		Hash []byte "json:\"hash\""
		Sig  []byte "json:\"sig\""
	}{
		Hash: hash,
		Sig:  sig,
	}
	return op, nil
}

func (bc *BlockChain) CurrentHeight() uint64 {
	bc.rwMutex.RLock()
	defer bc.rwMutex.RUnlock()
	return uint64(len(bc.blocks))
}

// TODO: USE ATOMIC
func (bc *BlockChain) setState(state MINING_STATE) {
	atomic.StoreInt32(bc.miningState.state, int32(state))
	fmt.Println("state changed to: ", state.String())
}

func (bc *BlockChain) readState() MINING_STATE {
	state := atomic.LoadInt32(bc.miningState.state)
	return MINING_STATE(state)
}

func (bc *BlockChain) pauseMining() {
	atomic.StoreInt32(bc.miningState.miningPaused, 1)
}

func (bc *BlockChain) resumeMining() {
	atomic.StoreInt32(bc.miningState.miningPaused, 0)
}

func (bc *BlockChain) isMiningPaused() bool {
	return atomic.LoadInt32(bc.miningState.miningPaused) == 1
}

func (bc *BlockChain) startOpMiningTimer() {
	if bc.miningState.startOpMiningTimerRunning.Load() == 1 {
		return
	}

	// lock the start op mining timer
	bc.miningState.startOpMiningTimerRunning.Store(1)

	go func() {
		bc.miningState.timeStartChan <- struct{}{}
		fmt.Println("starting timer")

		if bc.readState() == NO_OPERATION_MINING {
			bc.abandonMining()
		}

		// wait for state to change
		bc.waitTillStateChange(WAITING_FOR_OPERATION)

		bc.miningState.startOpMiningTimerRunning.Store(0)

		<-time.After(GenOpBlockTimeout)

		bc.miningState.timeOutChan <- struct{}{}
	}()
}

func (bc *BlockChain) abandonMining() {
	bc.miningState.abandonMining <- struct{}{}
}

func (bc *BlockChain) waitTillStateChange(state MINING_STATE) {
	for {
		if bc.readState() == state {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (bc *BlockChain) addBlockCleanup() {
	// correct our nonce
	// bc.Wallet.SetNonce(bc.ledger.GetAccount(Address(bc.Wallet.Address)).GetNonce())
}

func (bc *BlockChain) GetBlock(height uint64) (Block, error) {
	if height >= bc.CurrentHeight() {
		return Block{}, errors.New("block height out of range")
	}
	return bc.blocks[height-1], nil
}

func (bc *BlockChain) GetFullBlockchain() []Block {
	bc.rwMutex.RLock()
	defer bc.rwMutex.RUnlock()
	return bc.blocks
}

func (bc *BlockChain) GetMemPool() map[string]OperationMsg {
	bc.rwMutex.RLock()
	defer bc.rwMutex.RUnlock()
	return bc.operationMemPool.operationMemPool
}
