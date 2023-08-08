package blockchain

import (
	"fmt"
	"sync"
)

type OperationMemPool struct {
	mu               sync.RWMutex
	operationMemPool map[string]OperationMsg
	ledgerCopy       *LedgerStore
	// TODO: IMPLEMENT RFS COPY
	rfsCopy *RFSStore
}

func NewOperationMemPool(ledgerCopy *LedgerStore) *OperationMemPool {
	return &OperationMemPool{
		operationMemPool: make(map[string]OperationMsg),
		ledgerCopy:       ledgerCopy,
	}
}

// add operation to mempool & apply to ledger copy to see if all operations in mempool are valid
func (m *OperationMemPool) AddOperation(op OperationMsg) error {
	if op.Hash == "" {
		op.GenerateHash()
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.operationMemPool[op.Hash]; ok {
		return ErrAlreadyInMempool
	}

	// validate operation
	if ok, err := ValidateOperation(op, op.OpFrom, m.ledgerCopy.GetAccount(op.OpFrom)); !ok {
		return err
	}

	// apply operation to our ledger copy
	err := m.ledgerCopy.ApplyOperation(op)
	fmt.Println("balance left: ", m.ledgerCopy.GetAccount(op.OpFrom).GetBalance())
	if err != nil {
		return err
	}

	m.operationMemPool[op.Hash] = op
	return nil
}

// get copy of mempool
func (m *OperationMemPool) EmptyMemPool() map[string]OperationMsg {
	m.mu.RLock()
	defer m.mu.RUnlock()

	new := make(map[string]OperationMsg)

	for k, v := range m.operationMemPool {
		new[k] = v
	}

	m.operationMemPool = make(map[string]OperationMsg)
	// TODO: rfs copy

	return new
}

// check if operation exists in mempool
func (m *OperationMemPool) Exists(op OperationMsg) bool {
	if op.Hash == "" {
		op.GenerateHash()
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.operationMemPool[op.Hash]
	return ok
}

// remove operation from mempool, quiet
func (m *OperationMemPool) RemoveQuiet(op OperationMsg) {
	if op.Hash == "" {
		op.GenerateHash()
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.operationMemPool, op.Hash)
}

func (m *OperationMemPool) SetLedger(ledger *LedgerStore) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ledgerCopy = ledger
}

func (m *OperationMemPool) Length() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.operationMemPool)
}
