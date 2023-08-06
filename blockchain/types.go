package blockchain

import (
	"math"
	"sync"
	"time"
)

type Address string

const (
	MinedCoinsPerOpBlock   uint8         = 10
	MinedCoinsPerNoOpBlock uint8         = 5
	NumCoinsPerFileCreate  uint8         = 5
	GenOpBlockTimeout      time.Duration = time.Second * 10
	GenesisBlockHash       string        = "c6c534e825f4a3d41ede3e67473187d1"
	Difficulty             uint64        = math.MaxUint64
)

type RFSStore struct {
	rwMutex sync.RWMutex    `json:"-"`
	Files   map[string]File `json:"files"`
}

type LedgerStore struct {
	rwMutex sync.RWMutex               `json:"-"`
	Ledger  map[Address]AccountStorage `json:"ledger"`
}

type AccountStorage struct {
	rwMutex sync.RWMutex `json:"-"`
	Balance uint64       `json:"balance"`
	Nonce   uint64       `json:"nonce"`
}

type Record [512]byte

type File struct {
	Fname string `json:"fname"`
	// max number of records is 65535
	Records []Record `json:"records"`
}
