package blockchain

import (
	"context"
	"fmt"
	"sync"
)

func (bc *BlockChain) blockPubSubWorker() {
	for {
		select {
		case <-bc.ctx.Done():
			return
		case block := <-bc.blockListener:
			if bc.CurrentHeight() >= block.Height {
				continue
			}
			if bc.isMiningPaused() {
				continue
			}
			bc.blockPublisher <- block
			bc.AddLatestBlock(block)
		}
	}
}

func (bc *BlockChain) operationPubSubWorker() {
	for {
		select {
		case <-bc.ctx.Done():
			return
		case op := <-bc.operationListener:
			if bc.operationMemPool.Exists(op) {
				continue
			}
			if bc.isMiningPaused() {
				continue
			}
			bc.operationPublisher <- op
			bc.AddOperation(op)
		}
	}
}

func (bc *BlockChain) minerWorker() {

	var noOpCtx, opCtx context.Context
	var noOpCancel, opCancel context.CancelFunc

	noOpCtx, noOpCancel = context.WithCancel(bc.ctx)
	opCtx, opCancel = context.WithCancel(bc.ctx)

	go func() {
		for {
			select {
			case <-bc.ctx.Done():
				return
			case <-bc.miningState.abandonMining:
				fmt.Println("abandoning mining")
				noOpCancel()
				opCancel()
			}
		}
	}()

	for {
		fmt.Printf("Blockchain Current Height: %v \n", bc.CurrentHeight())
		if bc.isMiningPaused() {
			continue
		}

		select {
		case <-bc.ctx.Done():
			return
		case <-bc.miningState.timeStartChan:
			opCtx, opCancel = context.WithCancel(bc.ctx)
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
							bc.blockMined(opBlock)
						}
						return
					}

				}
			}(opCtx)
			wg.Wait()
		default:
			noOpCtx, noOpCancel = context.WithCancel(context.Background())
			if bc.operationMemPool.Length() > 0 {
				continue
			}
			func() {
				// mine no-op block
				noOpBlock := bc.InitNoOpBlock()
				defer bc.setState(IDLE)
				bc.setState(NO_OPERATION_MINING)
				noOpBlock, err := MineBlock(noOpBlock, Difficulty, noOpCtx.Done())
				if err != nil {
					return
				}
				bc.blockMined(noOpBlock)
			}()
		}
	}
}
