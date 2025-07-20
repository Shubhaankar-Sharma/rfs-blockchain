package blockchain

type Network interface {
	// GetBestBlockchain() ([]Block, map[string]OperationMsg, error)
	GetBestHeight() (uint64, string)
	PublishOperation(OperationMsg)
	SyncBlockchain() ([]Block, map[string]OperationMsg, error)
	BlockMined(Block)
	// GetBlockByNumber(uint64) (Block, error)
	// NumberOfPeers() int
}
