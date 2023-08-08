package blockchain

import "errors"

var (
	ErrInvalidBlock      = errors.New("invalid block")
	ErrAlreadyInMempool  = errors.New("already in mempool")
	ErrInvalidOperation  = errors.New("invalid operation")
	ErrInsufficientFunds = errors.New("insufficient funds")
)
