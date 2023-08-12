package blockchain

type Network interface {
	// GetBestBlockchain() ([]Block, map[string]OperationMsg, error)
	GetBestHeight() (uint64, string)
	// GetBlockByNumber(uint64) (Block, error)
	// NumberOfPeers() int
}
