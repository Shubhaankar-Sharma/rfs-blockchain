package blockchain

import (
	"sync"
	"time"
)

type Address string

const (
	MinedCoinsPerOpBlock   uint8         = 10
	MinedCoinsPerNoOpBlock uint8         = 5
	NumCoinsPerFileCreate  uint8         = 5
	NumCoinsPerFileAppend  uint8         = 5
	NumCoinsPerFileDelete  uint8         = 5
	GenOpBlockTimeout      time.Duration = time.Second * 10
	GenesisBlockHash       string        = "c6c534e825f4a3d41ede3e67473187d1"
	// ideal difficulty for a public blockchain would be 150 or above
	Difficulty uint64 = 151
)

type RFSStore struct {
	rwMutex sync.RWMutex    `json:"-"`
	Files   map[string]File `json:"files"`
}

type Record [512]byte

type File struct {
	Fname string `json:"fname"`
	// max number of records is 65535
	Records []Record `json:"records"`
}
