package blockchain

import (
	"errors"
	"sync"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
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

	ls.Ledger[address].Balance += amount
}

func (ls *LedgerStore) SubtractBalance(address Address, amount uint64) error {
	ls.rwMutex.Lock()
	defer ls.rwMutex.Unlock()
	if ls.Ledger[address] == nil {
		ls.Ledger[address] = BlankAccount()
		return ErrInsufficientFunds
	}

	if ls.Ledger[address].Balance < amount {
		return ErrInsufficientFunds
	}

	ls.Ledger[address].Balance -= amount
	return nil
}

func (ls *LedgerStore) IncrementNonce(address Address) {
	ls.rwMutex.Lock()
	defer ls.rwMutex.Unlock()
	if ls.Ledger[address] == nil {
		ls.Ledger[address] = BlankAccount()
	}

	ls.Ledger[address].Nonce += 1
}

func (ls *LedgerStore) NewTransaction() *Transaction {
	return &Transaction{
		Original:   ls,
		LedgerCopy: ls.copy(),
	}
}

func (ls *LedgerStore) ApplyOperation(op OperationMsg) error {
	switch op.OpType {
	case CREATE_FILE:
		ls.SubtractBalance(op.OpFrom, uint64(NumCoinsPerFileCreate))
	case APPEND_RECORD:
		ls.SubtractBalance(op.OpFrom, uint64(NumCoinsPerFileAppend))
	case DELETE_FILE:
		ls.SubtractBalance(op.OpFrom, uint64(NumCoinsPerFileDelete))
	}

	ls.IncrementNonce(op.OpFrom)
	return nil
}

func (ls *LedgerStore) copy() *LedgerStore {
	l := NewLedgerStore()
	ls.rwMutex.RLock()
	for k, v := range ls.Ledger {
		l.Ledger[k] = &AccountStorage{
			Balance: v.Balance,
			Nonce:   v.Nonce,
		}
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

func BlankAccount() *AccountStorage {
	return &AccountStorage{
		Balance: 0,
		Nonce:   0,
	}
}
