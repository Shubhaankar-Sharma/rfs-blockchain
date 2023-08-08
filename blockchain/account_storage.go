package blockchain

import "sync/atomic"

type AccountStorage struct {
	Balance *uint64 `json:"balance"`
}

func BlankAccount() *AccountStorage {
	b := uint64(0)
	return &AccountStorage{
		Balance: &b,
	}
}

func (a *AccountStorage) GetBalance() uint64 {
	return atomic.LoadUint64(a.Balance)
}

func (a *AccountStorage) SetBalance(b uint64) {
	atomic.StoreUint64(a.Balance, b)
}

func (a *AccountStorage) AddBalance(b uint64) {
	atomic.AddUint64(a.Balance, b)
}

func (a *AccountStorage) SubtractBalance(b uint64) error {
	if a.GetBalance() < b {
		return ErrInsufficientFunds
	}

	atomic.AddUint64(a.Balance, ^uint64(b-1))
	return nil
}

func (a *AccountStorage) NewCopy() *AccountStorage {
	b := a.GetBalance()
	a = BlankAccount()
	a.SetBalance(b)
	return a
}
