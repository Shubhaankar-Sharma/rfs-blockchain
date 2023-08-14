package network

type Operation uint8

const (
	// Commands
	UNKOWN Operation = iota
	GET_FULL_BLOCKCHAIN
	GET_BLOCK
	GET_MEM_POOL
	GET_BEST_HEIGHT
	GET_PEERS
	BLOCK_MINED
	SEND_BLOCK
	SEND_FULL_BLOCKCHAIN
	SEND_OPERATION
	SEND_MEM_POOL
	SEND_HEIGHT
	SEND_PEERS
)

func IsLazyOperation(op Operation) bool {
	switch op {
	case GET_FULL_BLOCKCHAIN, GET_MEM_POOL, GET_PEERS:
		return false
	case SEND_FULL_BLOCKCHAIN, SEND_MEM_POOL, SEND_PEERS:
		return false
	case GET_BEST_HEIGHT, SEND_OPERATION, SEND_HEIGHT, BLOCK_MINED, GET_BLOCK, SEND_BLOCK:
		return true
	}
	return false
}
