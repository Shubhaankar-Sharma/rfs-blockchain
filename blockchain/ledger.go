package blockchain

import (
	"errors"
	"sync"
)

type LedgerStore struct {
	rwMutex sync.RWMutex                `json:"-"`
	Ledger  map[Address]*AccountStorage `json:"ledger"`
}

type Transaction struct {
	Original   *LedgerStore
	LedgerCopy *LedgerStore
}

func NewLedgerStore() *LedgerStore {
	return &LedgerStore{
		Ledger: make(map[Address]*AccountStorage),
	}
}

func (ls *LedgerStore) GetAccount(address Address) *AccountStorage {
	ls.rwMutex.RLock()
	account := ls.Ledger[address]
	ls.rwMutex.RUnlock()
	if account == nil {
		return ls.CreateAccount(address, BlankAccount())
	}
	return account
}

func (ls *LedgerStore) CreateAccount(address Address, account *AccountStorage) *AccountStorage {
	ls.rwMutex.Lock()
	ls.Ledger[address] = account
	ls.rwMutex.Unlock()
	return account
}

func (ls *LedgerStore) AddBalance(address Address, amount uint64) {
	ls.rwMutex.Lock()
	defer ls.rwMutex.Unlock()
	if ls.Ledger[address] == nil {
		ls.Ledger[address] = BlankAccount()
	}

	ls.Ledger[address].AddBalance(amount)
}

func (ls *LedgerStore) SubtractBalance(address Address, amount uint64) error {
	ls.rwMutex.Lock()
	defer ls.rwMutex.Unlock()
	if ls.Ledger[address] == nil {
		ls.Ledger[address] = BlankAccount()
		return ErrInsufficientFunds
	}

	if ls.Ledger[address].GetBalance() < amount {
		return ErrInsufficientFunds
	}

	ls.Ledger[address].SubtractBalance(amount)
	return nil
}

func (ls *LedgerStore) NewTransaction() *Transaction {
	return &Transaction{
		Original:   ls,
		LedgerCopy: ls.copy(),
	}
}

func (ls *LedgerStore) ApplyOperation(op OperationMsg) error {
	var err error
	switch op.OpType {
	case CREATE_FILE:
		err = ls.SubtractBalance(op.OpFrom, uint64(NumCoinsPerFileCreate))
	case APPEND_RECORD:
		err = ls.SubtractBalance(op.OpFrom, uint64(NumCoinsPerFileAppend))
	case DELETE_FILE:
		err = ls.SubtractBalance(op.OpFrom, uint64(NumCoinsPerFileDelete))
	default:
		err = errors.New("invalid operation type")
	}

	if err != nil {
		return err
	}

	return nil
}

func (ls *LedgerStore) copy() *LedgerStore {
	l := NewLedgerStore()
	ls.rwMutex.RLock()
	for k, v := range ls.Ledger {
		l.Ledger[k] = v.NewCopy()
	}
	ls.rwMutex.RUnlock()
	return l
}

func (tx *Transaction) Commit() {
	tx.Original.rwMutex.Lock()
	tx.LedgerCopy.rwMutex.RLock()
	defer tx.Original.rwMutex.Unlock()
	defer tx.LedgerCopy.rwMutex.RUnlock()
	tx.Original.Ledger = tx.LedgerCopy.Ledger
}
